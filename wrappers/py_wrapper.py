import json
import sys
import importlib.util

def main():
  input_data = json.load(sys.stdin)

  spec = importlib.util.spec_from_file_location("module", "index.py")
  module = importlib.util.module_from_spec(spec)
  spec.loader.exec_module(module)

  if hasattr(module, "handler"):
    context = input_data["context"]
    data = input_data["data"]
    result = module.handler(context, data)
    print(result)
  else:
    print("cannot find handler function in index file")

if __name__ == "__main__":
  main()