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
 *    Andrew Young
 */

package mqtt

import (
	"net/url"
	"testing"
)

func Test_UsernamePassword(t *testing.T) {
	options := NewClientOptions()
	options.Username = "username"
	options.Password = "password"

	m := newConnectMsgFromOptions(options, &url.URL{})

	if m.Username != "username" {
		t.Fatalf("Username not set correctly")
	}

	if string(m.Password) != "password" {
		t.Fatalf("Password not set correctly")
	}
}

func Test_CredentialsProvider(t *testing.T) {
	options := NewClientOptions()
	options.Username = "incorrect"
	options.Password = "incorrect"
	options.SetCredentialsProvider(func() (username string, password string) {
		return "username", "password"
	})

	m := newConnectMsgFromOptions(options, &url.URL{})

	if m.Username != "username" {
		t.Fatalf("Username not set correctly")
	}

	if string(m.Password) != "password" {
		t.Fatalf("Password not set correctly")
	}
}

func Test_BrokerCredentials(t *testing.T) {
	m := newConnectMsgFromOptions(
		NewClientOptions(),
		&url.URL{User: url.UserPassword("username", "password")},
	)
	if m.Username != "username" {
		t.Fatalf("Username not set correctly")
	}
	if string(m.Password) != "password" {
		t.Fatalf("Password not set correctly")
	}
}
