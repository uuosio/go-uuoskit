package main

type TransactionExtension struct {
	Type uint16
	Data []byte
}

func (a *TransactionExtension) EstimatePackedSize() int {
	return 2 + 5 + len(a.Data)
}

func (t *TransactionExtension) Pack() []byte {
	enc := NewEncoder(2 + 5 + len(t.Data))
	enc.Pack(t.Type)
	enc.Pack(t.Data)
	return enc.GetBytes()
}

func (t *TransactionExtension) Unpack(data []byte) (int, error) {
	var err error
	dec := NewDecoder(data)
	t.Type, err = dec.UnpackUint16()
	if err != nil {
		return 0, err
	}

	t.Data, err = dec.UnpackBytes()
	if err != nil {
		return 0, err
	}
	return dec.Pos(), nil
}

type Transaction struct {
	// time_point_sec  expiration;
	// uint16_t        ref_block_num;
	// uint32_t        ref_block_prefix;
	// unsigned_int    max_net_usage_words = 0UL; /// number of 8 byte words this transaction can serialize into after compressions
	// uint8_t         max_cpu_usage_ms = 0UL; /// number of CPU usage units to bill transaction for
	// unsigned_int    delay_sec = 0UL; /// number of seconds to delay transaction, default: 0
	Expiration     uint32
	RefBlockNum    uint16
	RefBlockPrefix uint32
	//[VLQ or Base-128 encoding](https://en.wikipedia.org/wiki/Variable-length_quantity)
	//unsigned_int vaint (eosio.cdt/libraries/eosiolib/core/eosio/varint.hpp)
	MaxNetUsageWords   uint32
	MaxCpuUsageMs      uint8
	DelaySec           uint32 //unsigned_int
	ContextFreeActions []Action
	Actions            []Action
	Extention          []TransactionExtension
}

func NewTransaction(expiration int, taposBlockNum int, taposBlockPrefix int, delaySec int) *Transaction {
	t := &Transaction{}
	t.Expiration = uint32(expiration)
	t.RefBlockNum = uint16(taposBlockNum)
	t.RefBlockPrefix = uint32(taposBlockPrefix)
	t.MaxNetUsageWords = uint32(0)
	t.MaxCpuUsageMs = uint8(0)
	t.DelaySec = uint32(delaySec)
	return t
}

func (t *Transaction) Pack() []byte {
	initSize := 4 + 2 + 4 + 5 + 1 + 5

	initSize += 5 // Max varint size
	for _, action := range t.ContextFreeActions {
		initSize += action.EstimatePackedSize()
	}

	initSize += 5 // Max varint size
	for _, action := range t.Actions {
		initSize += action.EstimatePackedSize()
	}

	initSize += 5 // Max varint size
	for _, extention := range t.Extention {
		initSize += extention.EstimatePackedSize()
	}
	enc := NewEncoder(initSize)
	enc.Pack(t.Expiration)
	enc.Pack(t.RefBlockNum)
	enc.Pack(t.RefBlockPrefix)
	enc.PackVarInt(t.MaxNetUsageWords)
	enc.PackUint8(t.MaxCpuUsageMs)
	enc.PackVarInt(t.DelaySec)

	enc.PackLength(len(t.ContextFreeActions))
	for _, action := range t.ContextFreeActions {
		enc.Pack(&action)
	}

	enc.PackLength(len(t.Actions))
	for _, action := range t.Actions {
		enc.Pack(&action)
	}

	enc.PackLength(len(t.Extention))
	for _, extention := range t.Extention {
		enc.Pack(&extention)
	}
	return enc.GetBytes()
}

func (t *Transaction) Unpack(data []byte) (int, error) {
	var err error

	dec := NewDecoder(data)
	t.Expiration, err = dec.UnpackUint32()
	if err != nil {
		return 0, err
	}

	t.RefBlockNum, err = dec.UnpackUint16()
	if err != nil {
		return 0, err
	}

	t.RefBlockPrefix, err = dec.UnpackUint32()
	if err != nil {
		return 0, err
	}

	t.MaxNetUsageWords, err = dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.MaxCpuUsageMs, err = dec.UnpackUint8()
	if err != nil {
		return 0, err
	}

	t.DelaySec, err = dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	contextFreeActionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.ContextFreeActions = make([]Action, contextFreeActionLength)
	for i := 0; i < int(contextFreeActionLength); i++ {
		_, err := dec.Unpack(&t.ContextFreeActions[i])
		if err != nil {
			return 0, err
		}
	}

	actionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}

	t.Actions = make([]Action, actionLength)
	for i := 0; i < int(actionLength); i++ {
		_, err := dec.Unpack(&t.Actions[i])
		if err != nil {
			return 0, err
		}
	}

	extentionLength, err := dec.UnpackVarInt()
	if err != nil {
		return 0, err
	}
	t.Extention = make([]TransactionExtension, extentionLength)
	for i := 0; i < int(extentionLength); i++ {
		t.Extention[i].Type, err = dec.UnpackUint16()
		if err != nil {
			return 0, err
		}

		t.Extention[i].Data, err = dec.UnpackBytes()
		if err != nil {
			return 0, err
		}
	}
	return dec.Pos(), nil
}
