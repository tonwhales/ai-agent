package main

type PllConst struct {
	Addr uint8
	Mask uint16
}

// PllConstValue PLL Constraint + Value
type PllConstValue struct {
	Const PllConst
	Value uint16
}

type XilinxProperty struct {
	PLLPowerAddr uint8
	PLLFreq      map[int][]PllConstValue
}

var (
	DivClk   = PllConst{0x16, 0xC000}
	ClkReg1  = PllConst{0x0A, 0x1000}
	ClkReg2  = PllConst{0x0B, 0xFC00}
	ClkFbOut = PllConst{0x14, 0x1000}
	FiltReg1 = PllConst{0x4E, 0x66FF}
	FiltReg2 = PllConst{0x4F, 0x666F}
	Lock1    = PllConst{0x18, 0xFC00}
	Lock2    = PllConst{0x19, 0x8000}
	Lock3    = PllConst{0x1A, 0x8000}

	ClkFbOut1 = PllConst{0x14, 0x1000}
	ClkFbOut2 = PllConst{0x15, 0x8000}
)

var (
	Xilinx7Series = XilinxProperty{
		PLLPowerAddr: 0x28,
		PLLFreq: map[int][]PllConstValue{
			50:  []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x03cf}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x8800}, {Lock1, 0x0145}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			60:  []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			70:  []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0555}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			80:  []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0618}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			90:  []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x06db}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			100: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0186}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x079e}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			110: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0145}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x06dc}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			120: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0145}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x079e}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			130: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0104}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x069a}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			140: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0104}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x071c}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			150: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0104}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x079e}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			160: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0410}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			170: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0451}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			180: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			190: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x04d3}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			200: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0514}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			210: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0555}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			220: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0596}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			230: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x05d7}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			240: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0618}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			250: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0659}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			260: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x069a}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			270: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x06db}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			280: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x071c}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			290: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x075d}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			300: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0082}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x079e}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			310: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x03d0}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			320: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0410}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			330: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0411}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			340: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0451}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			350: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0452}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			360: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			370: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0493}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			380: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x04d3}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			390: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x04d4}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			400: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0514}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			410: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0515}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			420: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0555}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			430: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0556}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			440: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0596}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			450: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0597}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			460: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x05d7}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			470: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x05d8}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			480: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0618}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			490: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0619}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			500: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x0659}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			510: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x065a}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			520: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x069a}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			530: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x069b}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			540: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x06db}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			550: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x06dc}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			560: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x071c}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			570: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x071d}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			580: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x075d}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			590: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x075e}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			600: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0000}, {ClkFbOut1, 0x079e}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			610: []PllConstValue{{DivClk, 0x0145}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x079f}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			620: []PllConstValue{{DivClk, 0x0145}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x07df}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			630: []PllConstValue{{DivClk, 0x0145}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x07e0}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			640: []PllConstValue{{DivClk, 0x0145}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0820}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x0800}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			650: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x03cf}, {ClkFbOut2, 0x4c00}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			660: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0411}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			670: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0410}, {ClkFbOut2, 0x4800}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x012c}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			680: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0451}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			690: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0410}, {ClkFbOut2, 0x4c00}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			700: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0452}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			710: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0451}, {ClkFbOut2, 0x4800}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			720: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			730: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0451}, {ClkFbOut2, 0x4c00}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x0113}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			740: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0493}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			750: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x4800}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			760: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x04d3}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			770: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0492}, {ClkFbOut2, 0x4c00}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			780: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x04d4}, {ClkFbOut2, 0x0080}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			790: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x04d3}, {ClkFbOut2, 0x4800}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
			800: []PllConstValue{{DivClk, 0x2083}, {ClkReg1, 0x0041}, {ClkReg2, 0x0040}, {ClkFbOut1, 0x0514}, {ClkFbOut2, 0x0000}, {FiltReg1, 0x0800}, {FiltReg2, 0x9000}, {Lock1, 0x00fa}, {Lock2, 0x7c01}, {Lock3, 0x7fe9}},
		},
	}
)
