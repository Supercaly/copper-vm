import subprocess
import os

build_dir = "build"
cmd_dir = "cmd"

executables = ["casm", "deasm", "emulator", "copperdb"]

if __name__ == "__main__":
   [subprocess.run([
       "go", "build",
       "-o", os.path.join(build_dir, ex),
       os.path.join(cmd_dir, ex, ex+".go")
   ]) for ex in executables]