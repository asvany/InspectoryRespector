#! /usr/bin/env python3

import sys
import ir_protocol_py as irp


def processFile(filename):
	#iterate throught the file and read as a protobuf entry from irp as an WindowChange message
	with open(filename, "rb") as f:
		print("Processing file: " + filename)
		while True:
			try:
				windowChange = irp.WindowChange()
				windowChange.ParseFromString(f.read())
				print(windowChange)
			except EOFError:
				break
			except Exception as e:
				print(e)
				break


def main():
	print("Hello World!")
	# parsing each of the command line arguments as a filename
	for filename in sys.argv[1:]:
		processFile(filename)
  
if __name__== "__main__":
	main()