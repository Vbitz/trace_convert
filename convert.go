package main

import (
	"bufio"
	"compress/gzip"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	pb "eait.uq.edu.au/jscarsbrook/traceconvert/v3/protos/protos/perfetto/trace"
	"google.golang.org/protobuf/proto"
)

var inputFilename = flag.String("input", "", "The extracted.gz file to read.")
var outputFilename = flag.String("output", "", "The .bin protobuf file to write.")

func main() {
	flag.Parse()

	if *inputFilename == "" || *outputFilename == "" {
		log.Fatalf("usage: convert_trace <input> <output>\n")
	}

	var events []*pb.FtraceEvent

	{
		f, err := os.Open(*inputFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		gr, err := gzip.NewReader(f)
		if err != nil {
			log.Fatal(err)
		}
		defer gr.Close()

		scanner := bufio.NewScanner(gr)

		for scanner.Scan() {
			tokens := strings.Split(scanner.Text(), " ")
			if tokens[0] == "#" {
				continue
			}
			ts, err := strconv.ParseUint(tokens[0], 10, 64)
			if err != nil {
				log.Fatal(err)
			}
			ts *= 10
			event := tokens[2]
			if event == "525" {
				ip := tokens[9]
				if strings.HasPrefix(ip, "enter=") {
					symbol := strings.TrimPrefix(ip, "enter=")
					if err != nil {
						log.Fatal(err)
					}
					buf := fmt.Sprintf("B|0|%v\n", symbol)
					events = append(events, &pb.FtraceEvent{
						Timestamp: &ts,
						Pid:       proto.Uint32(0),
						Event: &pb.FtraceEvent_Print{
							Print: &pb.PrintFtraceEvent{
								Ip:  proto.Uint64(0),
								Buf: &buf,
							},
						},
					})
					// log.Printf("ts=%d ip=%v", ts, ipInt)
				} else if strings.HasPrefix(ip, "leave=") {
					buf := "E|0\n"
					events = append(events, &pb.FtraceEvent{
						Timestamp: &ts,
						Pid:       proto.Uint32(0),
						Event: &pb.FtraceEvent_Print{
							Print: &pb.PrintFtraceEvent{
								Ip:  proto.Uint64(0),
								Buf: &buf,
							},
						},
					})
					// log.Printf("ts=%d ip=%v", ts, ipInt)
				}
			} else if event == "1280" {
				name := tokens[9]
				buf := fmt.Sprintf("B|0|%v\n", name)
				events = append(events, &pb.FtraceEvent{
					Timestamp: &ts,
					Pid:       proto.Uint32(0),
					Event: &pb.FtraceEvent_Print{
						Print: &pb.PrintFtraceEvent{
							Ip:  proto.Uint64(0),
							Buf: &buf,
						},
					},
				})
			} else if event == "1792" {
				buf := "E|0\n"
				events = append(events, &pb.FtraceEvent{
					Timestamp: &ts,
					Pid:       proto.Uint32(0),
					Event: &pb.FtraceEvent_Print{
						Print: &pb.PrintFtraceEvent{
							Ip:  proto.Uint64(0),
							Buf: &buf,
						},
					},
				})
			} else {
				// log.Printf("unhandled %v", tokens)
			}
		}
	}

	{
		trace := pb.Trace{
			Packet: []*pb.TracePacket{
				{Data: &pb.TracePacket_FtraceEvents{
					&pb.FtraceEventBundle{
						Cpu:   proto.Uint32(0),
						Event: events,
					}},
				},
			},
		}
		bytes, err := proto.Marshal(&trace)
		if err != nil {
			log.Fatal(err)
		}

		f, err := os.Create(*outputFilename)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		n, err := f.Write(bytes)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Wrote %v bytes", n)
	}
}
