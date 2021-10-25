import sys
import os
import subprocess

emulator_path = "./build/emulator"
casm_path = "./build/casm"

casm_file_ext = ".casm"
result_file_ext = ".txt"
bin_file_ext = ".copper"

def compile_folder(folder):
    assert os.path.isdir(folder)

    failed = 0
    files = list(filter(lambda p: p.endswith(casm_file_ext), os.listdir(folder)))
    for file in files:
        if not compile_file(os.path.join(folder, file)):
            failed+=1
    print()
    print("All tests compiled!")
    print(f"Passed: {len(files)-failed}, Failed: {failed}")
    if failed != 0:
        exit(1)

def compile_file(file_path):
    assert os.path.exists(file_path)
    assert os.path.isfile(file_path)

    out_file_path = os.path.splitext(file_path)[0] + bin_file_ext

    result = subprocess.run([
        casm_path, 
        "-t", "copper",
        "-o", out_file_path,
        file_path])
    if not result.returncode == 0:
        return False
    return True

def run_tests_for_dir(folder):
    assert os.path.isdir(folder)
    failed = 0
    files = list(filter(lambda p: p.endswith(casm_file_ext), os.listdir(folder)))
    for file in files:
        if not run_tests_for_file(os.path.join(folder, file)):
            failed += 1
    
    print()
    print("All tests executed!")
    print(f"Passed: {len(files)-failed}, Failed: {failed}")
    if failed != 0:
        exit(1)

def run_tests_for_file(file_path):
    result_file_path = os.path.splitext(file_path)[0] + result_file_ext
    bin_file_path = os.path.splitext(file_path)[0] + bin_file_ext

    if not os.path.exists(result_file_path):
        print(f"ERROR: file '{result_file_path}' doesn't exist")
        exit(1)
    if not os.path.exists(bin_file_path):
        print(f"ERROR: file '{bin_file_path}' doesn't exist")
        exit(1)

    with open(result_file_path, "rb") as result_file:
        expected_result = result_file.read()

    actual_result = subprocess.run([emulator_path, bin_file_path], capture_output=True)
    if actual_result.returncode != 0 or expected_result != actual_result.stdout:
        print(f"ERROR: test '{file_path}' returned a different result from expected")
        print(f"     status code: {actual_result.returncode}")
        print(f"     expected: {expected_result.decode('UTF-8')}")
        print(f"     actual: {actual_result.stdout.decode('UTF-8')}")
        return False
    return True

def record_folder_output(folder):
    assert os.path.isdir(folder)
    [record_file_output(os.path.join(folder, file)) for file in list(filter(lambda p: p.endswith(casm_file_ext), os.listdir(folder)))]

def record_file_output(file_path):
    out_file_path = os.path.splitext(file_path)[0] + result_file_ext
    bin_file_path = os.path.splitext(file_path)[0] + bin_file_ext

    if not os.path.exists(bin_file_path):
        print(f"ERROR: file '{bin_file_path}' doesn't exist")
        exit(1)

    test_result = subprocess.run([emulator_path, bin_file_path], capture_output=True)
    if test_result.returncode != 0:
        print(f"ERROR: cannot record test '{file_path}' because it failed")
        exit(1)

    with open(out_file_path, "wb+") as out_file:
        out_file.write(test_result.stdout)

def expect_arg(argv, program):
    if len(argv) == 0:
        usage(program)
        print("ERROR: expecting a file/folder path afer command")
        exit(1)

def usage(program):
    print(f"Usage {program} [OPRIONS] [SUBCOMMAND]")
    print("[USAGE]: ")
    print("    -h       Print this help message.")
    print("[SUBCOMMAND]:")
    print(f"    com    <folder|file{casm_file_ext}>     Compiles a given file or folder to binary")
    print(f"    run    <folder|file{casm_file_ext}>     Run all tests for a given file or folder")
    print(f"    record <folder|file{casm_file_ext}>     Record the outputs of given file or folder")

if __name__ == "__main__":
    program, *argv = sys.argv

    command = "run"
    if len(argv) > 0:
        command, *argv = argv

    if command == "-h":
        usage(program)
        exit(0)
    elif command == "com":
        expect_arg(argv, program)
        path, *argv = argv
        if os.path.isdir(path):
            compile_folder(path)
        else:
            ok = compile_file(path)
            print(f"Test {path} compiled with {'success' if ok else 'failure'}!")
            if not ok:
                exit(1)
    elif command == "run":
        expect_arg(argv, program)
        path, *argv = argv
        if os.path.isdir(path):
            run_tests_for_dir(path)
        else:
            ok = run_tests_for_file(path)
            print(f"Test {path} executed with {'success' if ok else 'failure'}!")
            if not ok:
                exit(1)
    elif command == "record":
        expect_arg(argv, program)
        path, *argv = argv
        if os.path.isdir(path):
            record_folder_output(path)
        else:
            record_file_output(path)
    else:
        usage(program)
        print(f"ERROR: unknown command {command}")
        exit(1)

