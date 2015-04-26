package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

func waitForNotification(l *pq.Listener) {
	for {
		select {
		case n := <-l.Notify:
			fmt.Printf("received notification, new work available %q\n", n.Extra)
			return
		case <-time.After(90 * time.Second):
			go func() {
				l.Ping()
			}()
			// Check if there's more work available, just in case it takes
			// a while for the Listener to notice connection loss and
			// reconnect.
			fmt.Println("received no work for 90 seconds, checking for new work")
			return
		}
	}
}

func main() {
	var conninfo string = "postgres://vagrant:vagrant@localhost/TD"

	_, err := sql.Open("postgres", conninfo)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err.Error())
		}
	}

	listener := pq.NewListener(conninfo, 10*time.Second, time.Minute, reportProblem)
	err = listener.Listen("sources")
	if err != nil {
		panic(err)
	}

	fmt.Println("entering main loop")
	for {
		// process all available work before waiting for notifications
		// getWork(db)
		waitForNotification(listener)
	}
}
