package chhash

import (
	"testing"
)

type Data struct {
	err     error
	raw     string
	encoded string
}

func TestPasswd(t *testing.T) {
	var datas = []Data{
		Data{nil, "123456", "413434637952514332487a465379594375324a446d6365596e43396e6d656670" +
			"70626b646632" +
			"fda50d2d74d57b42a02d9b2d5d09a69121eaef9394ee77f0f465fe0707bf12acf103823b522973f42ba50e52ab9ba99c8c4de514afd9746eba923984dc19ec8a"},
		Data{nil, "w456b3456b354tvq34vtq34tq34tvq34fvtq346tbq34btq34tb", "55414472675041577464385273475531716571356154613579434b6e56675732" +
			"70626b646632" +
			"1cb13f5bd159329f39cc824f198fc746b1206bae34b06066a7524f21cb3534144ee796873d21c691b3a6224c85ed90ea969c8a28a9a867163ac17171fb99b80c"},
		//wrong passwd
		Data{WrongPasswd, "w456b3456b354tvq34vtq34tq34tvq34fVtq346tbq34btq34tb", "55414472675041577464385273475531716571356154613579434b6e56675732" +
			"70626b646632" +
			"1cb13f5bd159329f39cc824f198fc746b1206bae34b06066a7524f21cb3534144ee796873d21c691b3a6224c85ed90ea969c8a28a9a867163ac17171fb99b80c"},

		//error encoded
		Data{WrongPasswd, "w456b3456b354tvq34vtq34tq34tvq34fvtq346tbq34btq34tb", "55414472675041577464385273475531716571356154613579434b6e56675732" +
			"70626b646632" +
			"1cb13f5bd159329f39cc824f198fc746b1206bbe34b06066a7524f21cb3534144ee796873d21c691b3a6224c85ed90ea969c8a28a9a867163ac17171fb99b80c"},
		// error length
		Data{WrongFormatPasswd, "w456b3456b354tvq34vtq34tq34tvq34fvtq346tbq34btq34tb", "655414472675041577464385273475531716571356154613579434b6e5667573270626b6466321cb13f5bd159329f39cc824f198fc746b1206bae34b06066a7524f21cb3534144ee796873d21c691b3a6224c85ed90ea969c8a28a9a867163ac17171fb99b80c"},
		// error pbkdf2
		Data{WrongFormatPasswd, "w456b3456b354tvq34vtq34tq34tvq34fvtq346tbq34btq34tb", "55414472675041577464385273475531716571356154613579434b6e56675732" +
			"70626b646633" +
			"1cb13f5bd159329f39cc824f198fc746b1206bae34b06066a7524f21cb3534144ee796873d21c691b3a6224c85ed90ea969c8a28a9a867163ac17171fb99b80c"},
	}
	for _, data := range datas {
		if err := VerifyPasswd(data.raw, data.encoded); data.err != err {
			t.Fatalf("%v expected but got %v", data.err, err)
		}
	}
}
