#!/usr/bin/env bash
set -euo pipefail

PACKAGE="${1:-./examples/hardware_test}"
UF2_FILE="/tmp/firmware.uf2"

echo "[1/4] Compiling ${PACKAGE} for nicenano..."
GOTOOLCHAIN=go1.24.0 tinygo build -target nicenano -o "${UF2_FILE}" "${PACKAGE}"

echo "[2/4] Connecting to bootloader..."

# Check if already mounted
MOUNT_DIR=$(lsblk -o MOUNTPOINTS -n /dev/sdb 2>/dev/null | awk '{print $1}' | grep -v '^$' || true)

if [[ -z "${MOUNT_DIR}" ]]; then
	if [[ -e /dev/ttyACM0 && ! -e /dev/sdb ]]; then
		echo "Pulsing 1200-baud reset on /dev/ttyACM0..."
		python3 -c "import serial; s = serial.Serial('/dev/ttyACM0', 1200); s.close()" 2>/dev/null || true
	fi

	# Wait up to 6 seconds for the SCSI removable disk to appear
	for i in {1..12}; do
		if [[ -e /dev/sdb || -e /dev/sdc ]]; then
			break
		fi
		sleep 0.5
	done

	# Mount whichever disk appeared
	for dev in /dev/sdb /dev/sdc; do
		if [[ -e "${dev}" ]]; then
			udisksctl mount -b "${dev}" >/dev/null 2>&1 || true
			MOUNT_DIR=$(lsblk -o MOUNTPOINTS -n "${dev}" 2>/dev/null | awk '{print $1}' | grep -v '^$' || true)
			if [[ -n "${MOUNT_DIR}" ]]; then
				break
			fi
		fi
	done
fi

if [[ -z "${MOUNT_DIR}" ]]; then
	MOUNT_DIR=$(find /run/media -maxdepth 2 -type d -name "NICENANO" 2>/dev/null | head -n 1 || true)
fi

if [[ -z "${MOUNT_DIR}" || ! -d "${MOUNT_DIR}" ]]; then
	echo "Error: NICENANO bootloader volume could not be located." >&2
	exit 1
fi

echo "[3/4] Writing firmware to ${MOUNT_DIR}..."
cp "${UF2_FILE}" "${MOUNT_DIR}/"
sync

echo "[4/4] Successfully flashed! Robot is now running ${PACKAGE}."
