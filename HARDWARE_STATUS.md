# GOtto Ninja Robot - Hardware Status & Handoff

**Last Updated**: October 10, 2026

## 1. Verified Hardware Pinout
- **Buzzer**: Header `031` (`machine.P0_31`) — Verified playing tones & Imperial March.
- **Right Foot**: Header `006` (`machine.P0_06`) — 360° continuous rotation (`PWM2`).
- **Right Leg**: Header `008` (`machine.P0_08`) — 180° hip tilt (`PWM1`).
- **Left Leg**: Header `017` (`machine.P0_17`) — 180° hip tilt (`PWM1`).
- **Left Foot**: Header `020` (`machine.P0_20`) — 360° continuous rotation (`PWM2`).
- **Ultrasonic Trigger**: Header `100` (`machine.P1_00`).
- **Ultrasonic Echo**: Header `011` (`machine.P0_11`).
- **Power Gate**: `machine.P0_13` (`board.EnableVcc()`).

> ⚠️ **Caution**: Pin `029` (`P0.29`) has a low-impedance ground path on this shield. Never drive it high as an output.

---

## 2. Power & Battery Status
- **Current State**: Battery holder wiring fixed; servos confirmed physically moving when powered via external battery + jumper set to `VIN`.
- **Why USB is currently needed**: The white connector has 3 pins `[ BAT | G | 5V ]`. With a 2-wire battery cable, `-` is on `G` and `+` is on `5V`. `BAT` (which routes to Nice!Nano pin 24 `RAW`) is unconnected.
- **Untethered Solution**: Bridge `BAT` to `5V` on the white connector (or connect a short USB cable from the 16340 shield's USB-A port to the Nice!Nano USB-C port). Both the MCU and servos will then run completely untethered on battery power.

---

## 3. Toolchain & Flashing Pipeline
- **Command**: `python3 scripts/flash.py <package>` (e.g. `python3 scripts/flash.py examples/walk_and_buzz`).
- **Permissions**: 100% unprivileged; uses udisksctl and `/etc/udev/rules.d/99-nicenano.rules`.
- **Environment**: Automatically manages `GOTOOLCHAIN=go1.24.0` for TinyGo 0.39.0 compatibility.

---

## 4. Serial Command Bridge to ESP32 Head
- **UART Peripheral**: `machine.UART0` configured at 115200 baud, 8N1.
- **RX Pin**: Header `022` (`machine.P0_22`) — receives motion commands from ESP32-S3 TX.
- **TX Pin**: Header `024` (`machine.P0_24`) — transmits status & proximity telemetry to ESP32-S3 RX.
- **Firmware Package**: [`examples/serial_bridge/main.go`](examples/serial_bridge/main.go) (Verified compiling with TinyGo 0.39.0).
- **Supported Motion Commands**: `WAKEUP`, `WALK <count>`, `MOONWALK <count>`, `PIVOT <turns>`, `WIGGLE`, `CURIOUS`, `SLEEP`, `STOP`.
- **Telemetry Emitted**: `READY` (on startup), `DONE <command>` (motion completion), `PROXIMITY <mm>` (when HC-SR04 detects target within 350mm).

---

## 5. Next Session Action Items
1. Reseat Nice!Nano into red shield and bridge `BAT` to `5V` (or use short USB cable from battery).
2. Flash `examples/serial_bridge` (`python3 scripts/flash.py examples/serial_bridge`).
3. Wire NiceNano Header 022 (`RX`) to ESP32 TX (`D4`), Header 024 (`TX`) to ESP32 RX (`D5`), and common GND.
4. Test live interrogation loop: verify citizen approach triggers `PROXIMITY` and ESP32 sends `WALK` / `WIGGLE` / `CURIOUS` commands.
