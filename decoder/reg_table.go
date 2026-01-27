package decoder

var regTable [8][2]RegisterInfo = [8][2]RegisterInfo{
	// w = 0          | w = 1
	{{RegisterA, 0, 1}, {RegisterA, 0, 2}},  // AL, AX
	{{RegisterC, 0, 1}, {RegisterC, 0, 2}},  // CL, CX
	{{RegisterD, 0, 1}, {RegisterD, 0, 2}},  // DL, DX
	{{RegisterB, 0, 1}, {RegisterB, 0, 2}},  // BL, BX
	{{RegisterA, 1, 1}, {RegisterSP, 0, 2}}, // AH, SP
	{{RegisterC, 1, 1}, {RegisterBP, 0, 2}}, // CH, BP
	{{RegisterD, 1, 1}, {RegisterSI, 0, 2}}, // DH, SI
	{{RegisterB, 1, 1}, {RegisterDI, 0, 2}}, // BH, DI
}

func getRegOperand(regField int, w int) RegisterOperand {
	result := RegisterOperand{
		Register: regTable[regField&0b111][w],
	}

	return result
}
