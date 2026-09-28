package main

import (
	"github.com/godbus/dbus/v5"
)

func notify(title, message string) error {
	conn, err := dbus.SessionBus()
	if err != nil {
		return err
	}
	defer func(conn *dbus.Conn) {
		_ = conn.Close()
	}(conn)

	obj := conn.Object(
		"org.freedesktop.Notifications",
		"/org/freedesktop/Notifications",
	)

	var id uint32

	err = obj.Call(
		"org.freedesktop.Notifications.Notify",
		0,
		"SshChamp",                // app_name
		uint32(0),                 // replaces_id
		"",                        // app_icon
		title,                     // summary
		message,                   // body
		[]string{},                // actions
		map[string]dbus.Variant{}, // hints
		int32(10000),              // expire_timeout (ms)
	).Store(&id)

	return err
}
