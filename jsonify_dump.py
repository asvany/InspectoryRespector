#! /usr/bin/env python3

import sys

from ir_protocol_py.ir_record_pb2 import WindowChange
from google.protobuf.json_format import MessageToJson

import json


def processFile(filename):
    events = []
    with open(filename, "rb") as f:
        print("Processing file: " + filename)
        while True:
            # read the length of the message
            length_bytes = f.read(4)
            if len(length_bytes) == 0:
                print("End of file")
                break
            elif len(length_bytes) < 4:
                print(
                    f"File is corrupted: (message length) read:{ len(length_bytes)} excepted:4",
                )
                break
            length = int.from_bytes(length_bytes, "little")
            # read the message
            data = f.read(length)
            if len(data) < length:
                print(f"File is corrupted (message) read:{len(data)} expected:{length}")
                break
            # convert the message from protobuff
            windowChange = WindowChange()
            windowChange.ParseFromString(data)
            # convert the protobuff message to json
            json_windowChange = MessageToJson(
                windowChange,
                including_default_value_fields=True,
                preserving_proto_field_name=True,
                indent="\t",
                sort_keys=True,
            )
            events.append(json_windowChange)
    print("no of events: ", len(events))
    return events


def main():
    print("Hello World!")
    # parsing each of the command line arguments as a filename
    for filename in sys.argv[1:]:
        events = processFile(filename)
        # concatenate all json string into one json string
        s = "[" + ",\n".join(events) + "]\n"
        with open(filename + ".json", "w") as f:
            f.write(s)


if __name__ == "__main__":
    main()
