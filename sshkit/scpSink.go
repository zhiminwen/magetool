package sshkit

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

func ack(w io.Writer) error {
	_, err := w.Write([]byte{'\x00'})
	return err
}

func nextSinkAction(buf *bufio.Reader) (byte, error) {
	b, err := buf.ReadByte()
	if err != nil {
		log.Printf("Failed to read sink protocol action:%v", err)
		return 0, err
	}

	switch b {
	case '\000':
		log.Printf("Acked")
	case '\001', '\002':
		line, _, err := buf.ReadLine()
		if err != nil {
			log.Printf("Failed to read error string:%v", err)
			return 0, fmt.Errorf("Sink protocol error. message unknown")
		}
		return 0, fmt.Errorf("Sink protocol error: %s", line)
	}

	return b, nil
}

func sinkProtocol(r io.Reader, w io.Writer, dstWriter io.Writer) error {
	//first send Ack()
	err := ack(w)
	if err != nil {
		log.Printf("Failed to kick off the starting ack:%v", err)
		return err
	}
	buf := bufio.NewReader(r)
	act, err := nextSinkAction(buf)

	var perm os.FileMode
	var size int64
	var filename string

	if act == 'C' {
		//file header
		_, err := fmt.Fscanf(buf, "%04o %d %s\n", &perm, &size, &filename)
		if err != nil {
			log.Printf("Failed to parse sink file header:%v", err)
			return err
		}
		// log.Printf("perm=%v,size=%d,filename=%s", perm, size, filename)

		err = ack(w)
		if err != nil {
			log.Printf("Failed to send ack for file header:%v", err)
			return err
		}
	} else {
		// not expecting any other act
		log.Printf("Unexpected action from sink. byte=%x", act)
		return fmt.Errorf("Unexpected action from sink")
	}

	iop := NewIOProgress(size, "Downloading", "Downloaded")
	teeReader := io.TeeReader(buf, iop)

	_, err = io.CopyN(dstWriter, teeReader, size)
	if err != nil {
		log.Printf("Failed to copy io: %v", err)
		return err
	}

	err = ack(w)
	if err != nil {
		log.Printf("Failed to send ack for file content completion:%v", err)
		return err
	}

	iop.FinalMessage()
	return nil
}
