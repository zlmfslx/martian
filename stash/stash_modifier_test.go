package stash

import (
	"net"
	"net/http"
	"testing"

	"github.com/google/martian/fifo"
	"github.com/google/martian/parse"
	"github.com/google/martian/port"
)

func TestStashRequest(t *testing.T) {
	fg := fifo.NewGroup()
	fg.AddRequestModifier(NewModifier("stashed-url"))
	pmod := port.NewModifier()
	pmod.UsePort(8080)
	fg.AddRequestModifier(pmod)

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	if err := fg.ModifyRequest(req); err != nil {
		t.Fatalf("smod.ModifyRequest(): got %v, want no error", err)
	}

	_, port, err := net.SplitHostPort(req.URL.Host)
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q): got %v, want no error", req.URL.Host, err)
	}

	if got, want := port, "8080"; got != want {
		t.Errorf("port: got %v, want %v", got, want)
	}

	if got, want := req.Header.Get("stashed-url"), "http://example.com"; got != want {
		t.Errorf("stashed-url header: got %v, want %v", got, want)
	}

}

func TestStashInvalidHeaderName(t *testing.T) {
	mod := NewModifier("invalid-chars-actually-work-;><@")

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	if err := mod.ModifyRequest(req); err != nil {
		t.Fatalf("smod.ModifyRequest(): got %v, want no error", err)
	}

	if got, want := req.Header.Get("invalid-chars-actually-work-;><@"), "http://example.com"; got != want {
		t.Errorf("stashed-url header: got %v, want %v", got, want)
	}
}

func TestModiferFromJSON(t *testing.T) {
	msg := []byte(`{
    "fifo.Group": {
      "scope": ["request", "response"],
      "modifiers": [
        {
          "stash.Modifier": {
            "scope": ["request"],
            "headerName": "stashed-url"
          }
        },
        {
          "port.Modifier": {
            "scope": ["request"],
            "port": 8080
          }
        }
      ]
    }
  }`)

	r, err := parse.FromJSON(msg)
	if err != nil {
		t.Fatalf("parse.FromJSON(): got %v, want no error", err)
	}

	reqmod := r.RequestModifier()
	if reqmod == nil {
		t.Fatal("reqmod: got nil, want not nil")
	}

	req, err := http.NewRequest("GET", "http://example.com", nil)
	if err != nil {
		t.Fatalf("NewRequest(): got %v, want no error", err)
	}

	if err := reqmod.ModifyRequest(req); err != nil {
		t.Fatalf("ModifyRequest(): got %v, want no error", err)
	}

	_, port, err := net.SplitHostPort(req.URL.Host)
	if err != nil {
		t.Fatalf("net.SplitHostPort(%q): got %v, want no error", req.URL.Host, err)
	}

	if got, want := port, "8080"; got != want {
		t.Errorf("port: got %v, want %v", got, want)
	}

	if got, want := req.Header.Get("stashed-url"), "http://example.com"; got != want {
		t.Errorf("stashed-url header: got %v, want %v", got, want)
	}
}

func TestModiferFromJSONInvalidConfigurations(t *testing.T) {
	msg := []byte(`{
      "stash.Modifier": {
        "scope": ["response"],
        "headerName": "stash-header"
      }
    }`)

	_, err := parse.FromJSON(msg)
	if err == nil {
		t.Fatalf("parseFromJSON(msg): Got no error, but should have gotten one.")
	}
}