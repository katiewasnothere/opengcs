package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/Microsoft/opengcs/internal/uevent"
	"github.com/mdlayher/netlink"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

func pollUevent(timeout uint64) error {
	ueventSock, err := uevent.NewConnection()
	if err != nil {
		return err
	}
	fmt.Println("connected to socket")

	duration := time.Duration(timeout) * time.Second
	timeoutChan := time.After(duration)
	fmt.Printf("timeout %v seconds\n", timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timed out after %v seconds", timeout)
		default:
			fmt.Println("about to read")

			_, data, err := ueventSock.ReadMsg()
			if err != nil {
				return err
			}
			msg := string(*data)
			fmt.Println(msg)

			event, err := uevent.Parse(data)
			if err != nil {
				return err
			}
			fmt.Println(event)
		}
	}

	return nil
}

func pollUeventMdlayher(timeout uint64) error {
	conn, err := netlink.Dial(unix.NETLINK_KOBJECT_UEVENT, &netlink.Config{
		Groups: 1,
	})
	if err != nil {
		return err
	}

	duration := time.Duration(timeout) * time.Second
	timeoutChan := time.After(duration)
	fmt.Printf("timeout %v seconds\n", timeout)

	for {
		select {
		case <-timeoutChan:
			return fmt.Errorf("timed out after %v seconds", timeout)
		default:
			fmt.Println("about to read")

			msgs, err := conn.Receive()
			if err != nil {
				return err
			}

			// mdlayher gets a raw syscall conn then tries to read raw
			// which involves using recvmsg directly
			// not sure if I need to read the messages raw
			// otherwise that code is just a nice error handling

			// I still need to parse them
			fmt.Println(msgs)
		}
	}

	return nil
}

func ueventMain() {
	timeout := flag.Uint64("timeout", 30, "timeout for reading uevents in seconds, default 30 seconds")
	flag.Parse()

	if err := pollUevent(*timeout); err != nil {
		logrus.Errorf("error in poll uevent: %s", err)
		os.Exit(-1)
	}
	os.Exit(0)
}
