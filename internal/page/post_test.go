package page

import (
	"testing"

	"hawx.me/code/assert"
)

func TestSyndicationurl(t *testing.T) {
	atProtoUrl := "at://did:plc:2n2izph6uhty5uhdx7l32p67/app.bsky.feed.post/3lrjay3eyla2q"
	httpsUrl := syndicationUrl(atProtoUrl, "rosshendry.com")

	assert.NotEqual(t, atProtoUrl, httpsUrl)
	assert.Equal(t, "https://bsky.app/profile/rosshendry.com/post/3lrjay3eyla2q", httpsUrl)
}
