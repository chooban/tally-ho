package webmention

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const waitTime = 5 * time.Millisecond

type mention struct {
	source string
	data   map[string][]interface{}
}

type fakeBlog struct {
	ch chan mention
}

func (b *fakeBlog) BaseURL() string {
	return "http://example.com/"
}

func (b *fakeBlog) Entry(url string) (map[string][]interface{}, error) {
	if url != "http://example.com/weblog/post-id" {
		return map[string][]interface{}{}, errors.New("what is that")
	}

	return map[string][]interface{}{}, nil
}

func (b *fakeBlog) Mention(source string, data map[string][]interface{}) error {
	b.ch <- mention{source, data}
	return nil
}

func stringHandler(s string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(s))
	}
}

func goneHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusGone)
	}
}

func sequenceHandlers(hs ...http.Handler) http.HandlerFunc {
	index := 0

	return func(w http.ResponseWriter, r *http.Request) {
		if index >= len(hs) {
			w.WriteHeader(999)
			return
		}

		hs[index].ServeHTTP(w, r)
		index++
	}
}

func newFormRequest(qs url.Values) *http.Request {
	req := httptest.NewRequest("POST", "http://localhost/", strings.NewReader(qs.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req
}

func TestMention(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWhenPostUpdated(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(sequenceHandlers(stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`), stringHandler(`
<div class="h-entry">
  <h1 class="p-name">A great reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`)))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}

	req = newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp = w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A great reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithHCardAndHEntry(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<div class="h-card">
  <p class="p-name">John Doe</p>
</div>

<div class="h-entry">
  <h1 class="p-name">A reply to some post</h1>
  <p>
    In <a class="u-in-reply-to" href="http://example.com/weblog/post-id">this post</a>, I disagree.
  </p>
</div>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"A reply to some post"},
			"in-reply-to": {"http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestLikeFromBlueskyAndBridgy(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
    <article class="h-entry">
        <span class="p-uid">tag:bsky.app,2013:at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrjabiusqo2g_liked_by_did:plc:2n2izph6uhty5uhdx7l32p67</span>
        
        <span class="p-author h-card">
            <data class="p-uid" value="tag:bsky.app,2013:did:plc:2n2izph6uhty5uhdx7l32p67"></data>
            <a class="p-name u-url" href="https://bsky.app/profile/rosshendry.com">Ross Hendry</a>
            <a class="u-url" href="https://rosshendry.com/"></a>
            <a class="u-url" href="https://tally-ho.fly.dev/"></a>
            <a class="u-url" href="https://bsky.app/profile/did:plc:2n2izph6uhty5uhdx7l32p67"></a>
            <span class="p-nickname">rosshendry.com</span>
            <img class="u-photo" src="https://cdn.bsky.app/img/avatar/plain/avatar.jpg" alt="Ross Hendry's avatar">
        </span>

        <a class="p-name u-url" href="https://bsky.app/profile/rosshendry.com/post/3lrjabiusqo2g#liked_by_did:plc:2n2izph6uhty5uhdx7l32p67"></a>
        
        <a class="u-like-of" href="https://bsky.app/profile/rosshendry.com/post/3lrjabiusqo2g"></a>
        <a class="u-like-of" href="http://example.com/weblog/post-id"></a>
    </article>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":      {""},
			"like-of":   {"https://bsky.app/profile/rosshendry.com/post/3lrjabiusqo2g", "http://example.com/weblog/post-id"},
			"hx-target": {"http://example.com/weblog/post-id"},
			"url":       {"https://bsky.app/profile/rosshendry.com/post/3lrjabiusqo2g#liked_by_did:plc:2n2izph6uhty5uhdx7l32p67"},
			"uid":       {"tag:bsky.app,2013:at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrjabiusqo2g_liked_by_did:plc:2n2izph6uhty5uhdx7l32p67"},
			"author": {map[string]interface{}{
				"name": "Ross Hendry",
				"properties": map[string]interface{}{
					"name":     "Ross Hendry",
					"url":      []interface{}{"https://bsky.app/profile/rosshendry.com", "https://rosshendry.com/", "https://tally-ho.fly.dev/", "https://bsky.app/profile/did:plc:2n2izph6uhty5uhdx7l32p67"},
					"nickname": []interface{}{"rosshendry.com"},
					"photo":    []interface{}{"https://cdn.bsky.app/img/avatar/plain/avatar.jpg"},
				},
			}},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}
func TestMentionFromBlueskyAndBridgy(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<article class="h-entry">
  <span class="p-uid">at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrl2lv4x2222</span>

  <time class="dt-published" datetime="2025-06-14T13:28:34.177000+00:00">2025-06-14T13:28:34.177000+00:00</time>

  <span class="p-author h-card">
    <data class="p-uid" value="did:plc:2n2izph6uhty5uhdx7l32p67"></data>
    <a class="p-name u-url" href="https://bsky.app/profile/rosshendry.com">Ross Hendry</a>
    <a class="u-url" href="https://rosshendry.com/"></a>
    <span class="p-nickname">rosshendry.com</span>
    <img class="u-photo" src="https://cdn.bsky.app/img/avatar/plain/did:plc:2n2izph6uhty5uhdx7l32p67/bafkreibxn3bwkijehc5dlhhv7cr3quzuy6ih6ygca4i26parzjxlayynui@jpeg" alt="" />
  </span>

  <a title="bsky.app/profile/rosshendry.com/post/3lrl2lv4x2222" class="u-url" href="https://bsky.app/profile/rosshendry.com/post/3lrl2lv4x2222">bsky.app/profile/rosshendry.com</a>

  <div class="e-content p-name">
    Another reply test
  </div>

  <a class="u-in-reply-to" href="at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrjabiusqo2g"></a>
  <a class="u-in-reply-to" href="https://bsky.app/profile/did:plc:2n2izph6uhty5uhdx7l32p67/post/3lrjabiusqo2g"></a>
  <a class="u-in-reply-to" href="http://example.com/weblog/post-id"></a>

</article>

`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"name":        {"Another reply test"},
			"in-reply-to": {"at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrjabiusqo2g", "https://bsky.app/profile/did:plc:2n2izph6uhty5uhdx7l32p67/post/3lrjabiusqo2g", "http://example.com/weblog/post-id"},
			"hx-target":   {"http://example.com/weblog/post-id"},
			"url":         {"https://bsky.app/profile/rosshendry.com/post/3lrl2lv4x2222"},
			"uid":         {"at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrl2lv4x2222"},
			"published":   {"2025-06-14T13:28:34.177000+00:00"},
			"author": {map[string]interface{}{
				"name": "Ross Hendry",
				"properties": map[string]interface{}{
					"name":     "Ross Hendry",
					"url":      []interface{}{"https://bsky.app/profile/rosshendry.com", "https://rosshendry.com/"},
					"nickname": []interface{}{"rosshendry.com"},
					"photo":    []interface{}{"https://cdn.bsky.app/img/avatar/plain/did:plc:2n2izph6uhty5uhdx7l32p67/bafkreibxn3bwkijehc5dlhhv7cr3quzuy6ih6ygca4i26parzjxlayynui@jpeg"},
				},
			}},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionWithoutMicroformats(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(stringHandler(`
<p>
  Just a link to <a href="http://example.com/weblog/post-id">this post</a>.
</p>
`))
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}

func TestMentionOfDeletedPost(t *testing.T) {
	assert := assert.New(t)

	blog := &fakeBlog{ch: make(chan mention, 1)}

	source := httptest.NewServer(goneHandler())
	defer source.Close()

	handler := Endpoint(blog)

	req := newFormRequest(url.Values{
		"source": {source.URL},
		"target": {"http://example.com/weblog/post-id"},
	})

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	resp := w.Result()
	assert.Equal(http.StatusAccepted, resp.StatusCode)

	select {
	case m := <-blog.ch:
		assert.Equal(source.URL, m.source)

		assert.Equal(map[string][]interface{}{
			"hx-target": {"http://example.com/weblog/post-id"},
			"hx-gone":   {true},
		}, m.data)
	case <-time.After(waitTime):
		t.Fatal("failed to get notified")
	}
}
