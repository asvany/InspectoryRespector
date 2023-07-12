#! /usr/bin/env python3

import sys

from ir_protocol_py.ir_record_pb2 import WindowChange
from google.protobuf.json_format import MessageToJson

import json

def processFile(filename):
	events=[]
	with open(filename, "rb") as f:
		print("Processing file: " + filename)
		n = 0
		while True:
			length_bytes = f.read(4)
			if len(length_bytes) < 4:
				print("End of file")
				break
			length = int.from_bytes(length_bytes, 'little')
			print("length: ", length)
			data = f.read(length)
			if len(data) < length:
				print("File is corrupted")
				break
			windowChange = WindowChange()
			windowChange.ParseFromString(data)
			n += 1
			#convert the protobuff message to json
			json_windowChange = MessageToJson(windowChange
                                     ,including_default_value_fields=True
                                     ,preserving_proto_field_name=True
                                     ,indent="\t"
                                     ,sort_keys=True
                                     )
			events.append(json_windowChange)
			print("no of events: ", n)
	return events

def main():
	print("Hello World!")
	# parsing each of the command line arguments as a filename
	for filename in sys.argv[1:]:
		events=processFile(filename)
		s="[\n"
		for event in events:
			s+=event+",\n"
		s=s[:-2]+"]\n"
		with open(filename+".json", "w") as f:
			f.write(s)
		
if __name__== "__main__":
	main()