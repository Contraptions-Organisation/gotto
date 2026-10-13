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

## 4. Next Session Action Items
1. Reseat Nice!Nano into red shield and bridge `BAT` to `5V` (or use short USB cable from battery).
2. Flash `examples/walk_and_buzz` and observe untethered biped walking.
3. Fine-tune trim balance (`ninja.Trim`) if needed.
4. Wire HW-486 / LM393 light sensor to expansion header (e.g. `022`).
