#!/usr/bin/env python3
import os
import shutil
import subprocess
import sys
import time

try:
    import serial
except ImportError:
    serial = None

package = sys.argv[1] if len(sys.argv) > 1 else './examples/hardware_test'
uf2_file = '/tmp/firmware.uf2'

print(f"[1/4] Compiling {package} for nicenano...")
res = subprocess.run(
    ['tinygo', 'build', '-target', 'nicenano', '-o', uf2_file, package],
    env=dict(os.environ, GOTOOLCHAIN='go1.24.0')
)
if res.returncode != 0:
    print("Compilation failed.", file=sys.stderr)
    sys.exit(res.returncode)

print("[2/4] Connecting to bootloader...")
mount_path = None
for path in ['/run/media/chris/NICENANO', '/media/chris/NICENANO']:
    if os.path.isdir(path):
        mount_path = path
        break

if not mount_path:
    if os.path.exists('/dev/ttyACM0') and serial:
        print("Pulsing 1200-baud reset on /dev/ttyACM0...")
        try:
            s = serial.Serial('/dev/ttyACM0', 1200)
            s.close()
        except Exception:
            pass
        time.sleep(2.0)

    for _ in range(25):
        for dev in ['/dev/sdb', '/dev/sdc']:
            if os.path.exists(dev):
                subprocess.run(['udisksctl', 'mount', '-b', dev], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                break
        for path in ['/run/media/chris/NICENANO', '/media/chris/NICENANO']:
            if os.path.isdir(path):
                mount_path = path
                break
        if mount_path:
            break
        time.sleep(0.3)

if not mount_path:
    print("Error: Could not locate or mount NICENANO volume.", file=sys.stderr)
    sys.exit(1)

print(f"[3/4] Writing firmware to {mount_path}...")
shutil.copyfile(uf2_file, os.path.join(mount_path, 'firmware.uf2'))
time.sleep(0.5)
print(f"[4/4] Successfully flashed! Robot is running {package}.")
