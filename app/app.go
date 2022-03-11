package app

import (
	"context"
	"github.com/eoscanada/eos-go"
	"recover_tool/conf"
	"sync"
)

var api *eos.API
var once sync.Once

func InitAPI() (err error) {

	once.Do(func() {

		api = eos.New(conf.APPConf().Node)

		keyBag := &eos.KeyBag{}
		err = keyBag.ImportPrivateKey(context.Background(), conf.APPConf().SendPrivateKey)
		if err != nil {

			return
		}

		api.SetSigner(keyBag)
	})

	return
}
