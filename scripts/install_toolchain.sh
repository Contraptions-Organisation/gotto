#!/usr/bin/env bash

set -euo pipefail

GO_VERSION=${GO_VERSION:-1.22.4}
TINYGO_VERSION=${TINYGO_VERSION:-0.40.1}
PROFILE_FILE="${HOME}/.profile"
SHELL_CONFIGS=("${HOME}/.profile" "${HOME}/.bashrc" "${HOME}/.bash_profile" "${HOME}/.zprofile" "${HOME}/.zshrc")
GO_COMPAT_MIN=${GO_COMPAT_MIN:-1.19}
GO_COMPAT_MAX=${GO_COMPAT_MAX:-1.22}
ALLOW_UNSUPPORTED_GO=${ALLOW_UNSUPPORTED_GO:-0}
GOTOOLCHAIN_VALUE="go${GO_VERSION}"
export GOTOOLCHAIN="${GOTOOLCHAIN_VALUE}"
OS_ID=""
OS_LIKE=""
SKIP_MICROPYTHON=${SKIP_MICROPYTHON:-0}
NEEDS_RELOGIN=0

log() {
	printf "[%s] %s\n" "$(date +'%H:%M:%S')" "$*"
}

SUDO_KEEPALIVE_PID=""

ensure_sudo() {
	if [[ "$(id -u)" -eq 0 ]]; then
		return
	fi
	if sudo -n true 2>/dev/null; then
		: # already have a cached/passwordless ticket
	elif [[ -t 0 || -t 1 ]]; then
		log "This installer needs sudo (package install, group membership, udev rules). Enter your password if prompted."
		if ! sudo -v; then
			echo "Could not obtain sudo access. Re-run this script from a real terminal, or as a user with sudo rights." >&2
			exit 1
		fi
	else
		cat >&2 <<-'EOF'
	This installer needs sudo, but no password can be prompted for in this
	non-interactive session (no TTY attached, e.g. run through an agent/CI).
	Run it from a real terminal instead, for example:

	    ./scripts/install_toolchain.sh

	or grant passwordless sudo for this user beforehand.
	EOF
		exit 1
	fi
	# Keep the sudo ticket alive for the duration of the script so later
	# `sudo` calls (apt, dpkg, usermod, ...) don't prompt again mid-run.
	( while true; do sudo -n true; sleep 60; kill -0 "$$" 2>/dev/null || exit; done ) &
	SUDO_KEEPALIVE_PID=$!
}

stop_sudo_keepalive() {
	if [[ -n "${SUDO_KEEPALIVE_PID}" ]]; then
		kill "${SUDO_KEEPALIVE_PID}" 2>/dev/null || true
	fi
}
trap stop_sudo_keepalive EXIT

append_path_line() {
	local line=$1
	touch "${PROFILE_FILE}"
	for cfg in "${SHELL_CONFIGS[@]}"; do
		[[ -f "${cfg}" ]] || continue
		if grep -qxF "${line}" "${cfg}"; then
			continue
		fi
		log "Adding PATH snippet to ${cfg}"
		echo "${line}" >>"${cfg}"
	done
}

version_minor_code() {
	local version=$1
	local major minor
	IFS='.' read -r major minor _ <<<"${version}"
	if [[ -z "${major}" || -z "${minor}" ]]; then
		echo 0
		return
	fi
	printf "%d%02d" "${major}" "${minor}"
}

ensure_go_version_supported() {
	if [[ "${ALLOW_UNSUPPORTED_GO}" == "1" ]]; then
		return
	fi
	local go_code min_code max_code
	go_code=$(version_minor_code "${GO_VERSION}")
	min_code=$(version_minor_code "${GO_COMPAT_MIN}")
	max_code=$(version_minor_code "${GO_COMPAT_MAX}")
	if (( go_code < min_code || go_code > max_code )); then
		cat >&2 <<-EOF
Go ${GO_VERSION} is outside the supported range (${GO_COMPAT_MIN} - ${GO_COMPAT_MAX}) required by TinyGo ${TINYGO_VERSION}.
Set GO_VERSION to a compatible release or export ALLOW_UNSUPPORTED_GO=1 to bypass this guard.
EOF
		exit 1
	fi
}

create_go_symlinks() {
	local go_bin="/usr/local/go/bin"
	if [[ ! -d "${go_bin}" ]]; then
		return
	fi
	log "Linking Go binaries into /usr/local/bin"
	for tool in go gofmt; do
		if [[ -x "${go_bin}/${tool}" ]]; then
			sudo ln -sf "${go_bin}/${tool}" "/usr/local/bin/${tool}"
		fi
	done
}

detect_os() {
	if [[ -f /etc/os-release ]]; then
		. /etc/os-release
		OS_ID=${ID:-unknown}
		OS_LIKE=${ID_LIKE:-}
	else
		OS_ID="unknown"
		OS_LIKE=""
	fi
}

install_packages() {
	local os_key="${OS_ID} ${OS_LIKE}"
	case "${os_key}" in
		*debian*|*ubuntu*|*pop*)
			log "Installing dependencies via apt"
			sudo apt update
			sudo apt install -y build-essential git wget tar clang gcc-arm-none-eabi libnewlib-arm-none-eabi python3 python3-pip python3-venv pipx
			;;
		*fedora*)
			log "Installing dependencies via dnf"
			sudo dnf install -y @'Development Tools' git wget tar clang arm-none-eabi-gcc-cs arm-none-eabi-newlib python3 python3-pip pipx
			;;
		*arch*|*manjaro*|*endeavouros*)
			log "Installing dependencies via pacman"
			sudo pacman -Sy --needed base-devel git wget tar clang arm-none-eabi-gcc arm-none-eabi-newlib python python-pip python-pipx
			;;
		unknown*)
			echo "Unsupported distribution: unable to detect OS. Please install dependencies manually." >&2
			exit 1
			;;
		*)
			echo "Unsupported distribution: ${OS_ID}. Please install dependencies manually." >&2
			exit 1
			;;
	esac
}

ensure_serial_permissions() {
	if ! getent group dialout >/dev/null 2>&1; then
		log "No 'dialout' group on this system; skipping serial permission setup"
		return
	fi
	if id -nG "${USER}" | tr ' ' '\n' | grep -qx dialout; then
		log "${USER} is already in the 'dialout' group"
		return
	fi
	log "Adding ${USER} to the 'dialout' group (required to flash/monitor over USB serial without sudo)"
	sudo usermod -aG dialout "${USER}"
	NEEDS_RELOGIN=1
}

install_micropython_tools() {
	if [[ "${SKIP_MICROPYTHON}" == "1" ]]; then
		log "Skipping MicroPython tooling (SKIP_MICROPYTHON=1)"
		return
	fi
	local tools=(mpremote esptool adafruit-nrfutil)
	append_path_line 'export PATH="$HOME/.local/bin:$PATH"'
	if command -v pipx >/dev/null 2>&1; then
		log "Installing MicroPython tooling (${tools[*]}) via pipx"
		for tool in "${tools[@]}"; do
			pipx install --force "${tool}"
		done
		pipx ensurepath >/dev/null 2>&1 || true
	elif command -v pip3 >/dev/null 2>&1; then
		log "pipx not available; installing MicroPython tooling (${tools[*]}) via pip3 --user"
		if ! pip3 install --user --upgrade "${tools[@]}" 2>/tmp/gotto-pip-error; then
			if grep -q externally-managed-environment /tmp/gotto-pip-error; then
				log "System pip is externally managed; retrying in an isolated venv at ${HOME}/.local/share/gotto-micropython"
				python3 -m venv "${HOME}/.local/share/gotto-micropython"
				"${HOME}/.local/share/gotto-micropython/bin/pip" install --upgrade "${tools[@]}"
				mkdir -p "${HOME}/.local/bin"
				for tool in "${tools[@]}"; do
					ln -sf "${HOME}/.local/share/gotto-micropython/bin/${tool}" "${HOME}/.local/bin/${tool}"
				done
			else
				cat /tmp/gotto-pip-error >&2
				rm -f /tmp/gotto-pip-error
				return 1
			fi
		fi
		rm -f /tmp/gotto-pip-error
	else
		log "Neither pipx nor pip3 found; skipping MicroPython tooling"
		return
	fi
	log "MicroPython tools installed: mpremote (REPL/file transfer), esptool (ESP32/8266 flashing), adafruit-nrfutil (nRF52 flashing/DFU)"
}

install_go() {
	ensure_go_version_supported
	local archive="/tmp/go${GO_VERSION}.linux-amd64.tar.gz"
	log "Downloading Go ${GO_VERSION}"
	wget -qO "${archive}" "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
	sudo rm -rf /usr/local/go
	log "Installing Go to /usr/local/go"
	sudo tar -C /usr/local -xzf "${archive}"
	append_path_line 'export PATH=/usr/local/go/bin:$PATH'
	append_path_line "export GOTOOLCHAIN=${GOTOOLCHAIN_VALUE}"
	export GOTOOLCHAIN="${GOTOOLCHAIN_VALUE}"
	create_go_symlinks
	log "$(/usr/local/go/bin/go version) installed"
}

install_tinygo_deb() {
	local deb="/tmp/tinygo_${TINYGO_VERSION}_amd64.deb"
	local url="https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/tinygo_${TINYGO_VERSION}_amd64.deb"
	log "Downloading TinyGo ${TINYGO_VERSION} .deb"
	wget -qO "${deb}" "${url}"
	log "Installing TinyGo via dpkg"
	if ! sudo dpkg -i "${deb}" >/dev/null; then
		log "Resolving TinyGo package dependencies"
		sudo apt install -f -y
		sudo dpkg -i "${deb}" >/dev/null
	fi
	rm -f "${deb}"
	log "$(tinygo version) installed"
}

install_tinygo_tar() {
	local archive="/tmp/tinygo${TINYGO_VERSION}.linux-amd64.tar.gz"
	log "Downloading TinyGo ${TINYGO_VERSION}"
	wget -qO "${archive}" "https://github.com/tinygo-org/tinygo/releases/download/v${TINYGO_VERSION}/tinygo${TINYGO_VERSION}.linux-amd64.tar.gz"
	sudo rm -rf /usr/local/tinygo
	log "Installing TinyGo to /usr/local/tinygo"
	sudo tar -C /usr/local -xzf "${archive}"
	append_path_line 'export PATH=/usr/local/tinygo/bin:$PATH'

	local tinygo_bin="/usr/local/tinygo/bin/tinygo"
	if [[ -x "${tinygo_bin}" ]]; then
		log "$("${tinygo_bin}" version) installed"
	else
		log "TinyGo extracted to /usr/local/tinygo"
	fi
	if [[ -d /usr/local/tinygo/udev ]]; then
		log "Applying TinyGo udev rules"
		sudo cp /usr/local/tinygo/udev/99-tinygo.rules /etc/udev/rules.d/
		sudo udevadm control --reload-rules
		sudo udevadm trigger
	fi
}

install_tinygo() {
	local os_key="${OS_ID} ${OS_LIKE}"
	case "${os_key}" in
		*debian*|*ubuntu*|*pop*)
			install_tinygo_deb
			;;
		*)
			install_tinygo_tar
			;;
	esac
}

main() {
	detect_os
	ensure_sudo
	log "Installing GOtto toolchain (Go ${GO_VERSION}, TinyGo ${TINYGO_VERSION})"
	install_packages
	install_go
	install_tinygo
	ensure_serial_permissions
	install_micropython_tools
	log "Installation complete. Restart your shell session or source your shell rc file (~/.profile, ~/.bashrc, ~/.zshrc) to refresh PATH."
	if [[ "${NEEDS_RELOGIN}" == "1" ]]; then
		log "You were just added to the 'dialout' group: log out and back in (or run 'newgrp dialout' in this shell) before flashing/monitoring over USB."
	fi
}

main "$@"
