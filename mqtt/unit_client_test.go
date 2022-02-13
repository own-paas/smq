/*
 * Copyright (c) 2021 IBM Corp and others.
 *
 * All rights reserved. This program and the accompanying materials
 * are made available under the terms of the Eclipse Public License v2.0
 * and Eclipse Distribution License v1.0 which accompany this distribution.
 *
 * The Eclipse Public License is available at
 *    https://www.eclipse.org/legal/epl-2.0/
 * and the Eclipse Distribution License is available at
 *   http://www.eclipse.org/org/documents/edl-v10.php.
 *
 * Contributors:
 *    Seth Hoenig
 *    Allan Stockdill-Mander
 *    Mike Robertson
 */

package mqtt

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"testing"
)

func init() {
	// Logging is off by default as this makes things simpler when you just want to confirm that tests pass
	// DEBUG = log.New(os.Stderr, "DEBUG    ", log.Ltime)
	// WARN = log.New(os.Stderr, "WARNING  ", log.Ltime)
	// CRITICAL = log.New(os.Stderr, "CRITICAL ", log.Ltime)
	// ERROR = log.New(os.Stderr, "ERROR    ", log.Ltime)

	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
}

func Test_NewClient_simple(t *testing.T) {
	ops := NewClientOptions().SetClientID("foo").AddBroker("tcp://10.10.0.1:1883")
	c := NewClient(ops).(*client)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	if c.options.ClientID != "foo" {
		t.Fatalf("bad client id")
	}

	if c.options.Servers[0].Scheme != "tcp" {
		t.Fatalf("bad server scheme")
	}

	if c.options.Servers[0].Host != "10.10.0.1:1883" {
		t.Fatalf("bad server host")
	}
}

func Test_NewClient_optionsReader(t *testing.T) {
	ops := NewClientOptions().SetClientID("foo").AddBroker("tcp://10.10.0.1:1883")
	c := NewClient(ops).(*client)

	if c == nil {
		t.Fatalf("ops is nil")
	}

	rOps := c.OptionsReader()
	cid := rOps.ClientID()

	if cid != "foo" {
		t.Fatalf("unable to read client ID")
	}

	servers := rOps.Servers()
	broker := servers[0]
	if broker.Hostname() != "10.10.0.1" {
		t.Fatalf("unable to read hostname")
	}

}

func Test_isConnection(t *testing.T) {
	ops := NewClientOptions()
	c := NewClient(ops)

	c.(*client).setConnected(connected)
	if !c.IsConnectionOpen() {
		t.Fail()
	}
}

func Test_isConnectionOpenNegative(t *testing.T) {
	ops := NewClientOptions()
	c := NewClient(ops)

	c.(*client).setConnected(reconnecting)
	if c.IsConnectionOpen() {
		t.Fail()
	}
	c.(*client).setConnected(connecting)
	if c.IsConnectionOpen() {
		t.Fail()
	}
	c.(*client).setConnected(disconnected)
	if c.IsConnectionOpen() {
		t.Fail()
	}
}
