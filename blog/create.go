package blog

import (
	"hawx.me/code/tally-ho/internal/mfutil"
)

func (b *Blog) Create(data map[string][]interface{}) (string, error) {
	b.massage(data)

	uid := mfutil.Get(data, "uid").(string)
	location := mfutil.Get(data, "url").(string)

	if err := b.entries.SetProperties(uid, data); err != nil {
		return location, err
	}

	go b.syndicate(location, data)
	go b.sendWebmentions(location, data)
	go b.hubPublish()

	return location, nil
}
