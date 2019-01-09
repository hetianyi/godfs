package bridgev2

import (
	"testing"
)

func TestServer(t *testing.T) {
	server := NewServer("", 1022)
	server.Listen()
}



func TestClient(t *testing.T) {

}