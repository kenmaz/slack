package slack

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

func postMessageInvalidChannelHandler(rw http.ResponseWriter, r *http.Request) {
	rw.Header().Set("Content-Type", "application/json")
	response, _ := json.Marshal(chatResponseFull{
		SlackResponse: SlackResponse{Ok: false, Error: "channel_not_found"},
	})
	rw.Write(response)
}

func TestPostMessageInvalidChannel(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	http.HandleFunc("/chat.postMessage", postMessageInvalidChannelHandler)
	once.Do(startServer)
	api := New("testing-token", OptionAPIURL("http://"+serverAddr+"/"))
	_, _, err := api.PostMessage("CXXXXXXXX", MsgOptionText("hello", false))
	if err == nil {
		t.Errorf("Expected error: channel_not_found; instead succeeded")
		return
	}

	if err.Error() != "channel_not_found" {
		t.Errorf("Expected error: channel_not_found; received: %s", err)
		return
	}
}

func TestGetPermalink(t *testing.T) {
	channel := "C1H9RESGA"
	timeStamp := "p135854651500008"

	http.HandleFunc("/chat.getPermalink", func(rw http.ResponseWriter, r *http.Request) {

		if got, want := r.Header.Get("Content-Type"), "application/x-www-form-urlencoded"; got != want {
			t.Errorf("request uses unexpected content type: got %s, want %s", got, want)
		}

		if got, want := r.URL.Query().Get("channel"), channel; got != want {
			t.Errorf("request contains unexpected channel: got %s, want %s", got, want)
		}

		if got, want := r.URL.Query().Get("message_ts"), timeStamp; got != want {
			t.Errorf("request contains unexpected message timestamp: got %s, want %s", got, want)
		}

		rw.Header().Set("Content-Type", "application/json")
		response := []byte("{\"ok\": true, \"channel\": \"" + channel + "\", \"permalink\": \"https://ghostbusters.slack.com/archives/" + channel + "/" + timeStamp + "\"}")
		rw.Write(response)
	})

	once.Do(startServer)
	api := New("testing-token", OptionAPIURL("http://"+serverAddr+"/"))
	pp := PermalinkParameters{Channel: channel, Ts: timeStamp}
	pl, err := api.GetPermalink(&pp)

	if got, want := pl, "https://ghostbusters.slack.com/archives/C1H9RESGA/p135854651500008"; got != want {
		t.Errorf("unexpected permalink: got %s, want %s", got, want)
	}

	if err != nil {
		t.Errorf("unexpected error returned: %v", err)
	}
}

func TestPostMessageWithBlocks(t *testing.T) {
	http.DefaultServeMux = new(http.ServeMux)
	http.HandleFunc("/chat.postMessage", func(rw http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		form, err := url.ParseQuery(string(body))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}
		actualBlocks := form.Get("blocks")
		expectedBlocks := `[{"type":"context","block_id":"context","elements":[{"type":"plain_text","text":"hello"}]}]`
		if expectedBlocks != actualBlocks {
			t.Errorf("expected: %s, got: %s", expectedBlocks, actualBlocks)
			return
		}
	})

	once.Do(startServer)
	api := New(validToken, OptionAPIURL("http://"+serverAddr+"/"))

	blocks := []Block{NewContextBlock("context", NewTextBlockObject(PlainTextType, "hello", false, false))}

	_, _, _ = api.PostMessage("CXXX", MsgOptionBlocks(blocks...), MsgOptionText("text", false))
}
