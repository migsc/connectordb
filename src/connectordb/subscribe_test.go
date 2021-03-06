/**
Copyright (c) 2016 The ConnectorDB Contributors
Licensed under the MIT license.
**/
package connectordb

import (
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/users"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscribe(t *testing.T) {
	Tdb.Clear()
	db := Tdb

	//Let's create a stream
	require.NoError(t, db.CreateUser(&users.UserMaker{User: users.User{Name: "tst", Email: "email@email", Password: "mypass", Role: "user", Public: true}}))
	require.NoError(t, db.CreateDevice("tst/tst", &users.DeviceMaker{}))
	require.NoError(t, db.CreateStream("tst/tst/tst", &users.StreamMaker{Stream: users.Stream{Schema: `{"type":"string"}`}}))

	recvchan := make(chan messenger.Message, 2)
	recvchan2 := make(chan messenger.Message, 2)
	recvchan3 := make(chan messenger.Message, 2)
	recvchan4 := make(chan messenger.Message, 2)
	//We bind a timeout to the channel, since we want the test to fail if no messages come through
	go func() {
		time.Sleep(2 * time.Second)
		recvchan <- messenger.Message{"TIMEOUT", "", []datastream.Datapoint{}}
		recvchan2 <- messenger.Message{"TIMEOUT", "", []datastream.Datapoint{}}
		recvchan3 <- messenger.Message{"TIMEOUT", "", []datastream.Datapoint{}}
		recvchan4 <- messenger.Message{"TIMEOUT", "", []datastream.Datapoint{}}
	}()

	_, err := db.Subscribe("tst", recvchan)
	require.NoError(t, err)
	_, err = db.Subscribe("tst/tst", recvchan2)
	require.NoError(t, err)
	_, err = db.Subscribe("tst/tst/tst", recvchan3)
	require.NoError(t, err)
	_, err = db.Subscribe("tst/tst/tst/downlink", recvchan4)
	require.NoError(t, err)
	db.Messenger.Flush() //Just to avoid problems

	data := []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 1.0,
		Data:      "Hello World!",
	}}

	require.NoError(t, db.InsertStream("tst/tst/tst", data, false))

	m := <-recvchan
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")
	m = <-recvchan2
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")
	m = <-recvchan3
	require.Equal(t, m.Stream, "tst/tst/tst")
	require.Equal(t, m.Data[0].Data, "Hello World!")

	data = []datastream.Datapoint{datastream.Datapoint{
		Timestamp: 2.0,
		Data:      "2",
	}}

	require.NoError(t, db.InsertStream("tst/tst/tst/downlink", data, false))
	db.Messenger.Flush()
	m = <-recvchan4
	require.Equal(t, m.Stream, "tst/tst/tst/downlink")
	require.Equal(t, m.Data[0].Data, "2")

	time.Sleep(100 * time.Millisecond)
	recvchan <- messenger.Message{"GOOD", "", []datastream.Datapoint{}}
	recvchan2 <- messenger.Message{"GOOD", "", []datastream.Datapoint{}}
	recvchan3 <- messenger.Message{"GOOD", "", []datastream.Datapoint{}}

	m = <-recvchan
	require.Equal(t, m.Stream, "GOOD", "A downlink should not be triggered")
	m = <-recvchan2
	require.Equal(t, m.Stream, "GOOD", "A downlink should not be triggered")
	m = <-recvchan3
	require.Equal(t, m.Stream, "GOOD", "A downlink should not be triggered")

}
