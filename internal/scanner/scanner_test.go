package scanner

import (
	"strings"
	"testing"

	"github.com/supaleon/vanilla/internal/token"
)

var fset = token.NewFileSet()

func TestAny(t *testing.T) {
	//fmt.Sprintf()
}

func TestHeaderChecking(t *testing.T) {
	tests := []struct {
		src, wantLit, wantErr string
		wantToken             token.Token
	}{
		{
			src:       " ",
			wantLit:   "", // whitespace will be removed.
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "🦍 ",
			wantLit:   "🦍 ",
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "< ",
			wantLit:   "< ",
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       " < ",
			wantLit:   "< ",
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "<🐶",
			wantLit:   "<🐶",
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "</🐶",
			wantLit:   "</",
			wantToken: token.ENDTagOpen,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "/>",
			wantLit:   "/>",
			wantToken: token.TEXT,
			wantErr:   "component source code must begin with a valid HTML tag",
		},
		{
			src:       "<div",
			wantLit:   "<",
			wantToken: token.STARTTagOpen,
			wantErr:   "",
		},
	}

	for _, test := range tests {
		var err string
		s := New(fset.AddFile("", fset.Base(), len(test.src)), []byte(test.src), func(_ token.Position, msg string) {
			if err == "" {
				err = msg
			}
		})
		_, tok, lit := s.Scan()
		switch tok {
		case token.STARTTagOpen:
			lit = "<"
		case token.ENDTagOpen:
			lit = "</"
		}
		if lit != test.wantLit {
			t.Errorf("%q: got literal %q (%s); want %s", test.src, lit, tok, test.wantToken)
		}

		if err != test.wantErr {
			t.Errorf("%q: got error %q; want %q", test.src, err, test.wantErr)
		}

		// make sure we read all
		s.Scan()
		_, end, _ := s.Scan()
		if end != token.EOF {
			t.Errorf("%q: got %s; want EOF", test.src, tok)
		}
	}
}

func TestNumbers(t *testing.T) {
	for _, test := range []struct {
		tok              token.Token
		src, tokens, err string
	}{
		// binaries
		{token.ILLEGAL, "0i", "0i", "imaginary numbers are not allowed"},
		{token.INT, "0b0", "0b0", ""},
		{token.INT, "0b1010", "0b1010", ""},
		{token.INT, "0B1110", "0B1110", ""},

		{token.INT, "0b", "0b", "binary literal has no digits"},
		{token.INT, "0b0190", "0b0190", "invalid digit '9' in binary literal"},
		{token.INT, "0b01a0", "0b01 a0", ""}, // only accept 0-9

		{token.FLOAT, "0b.", "0b.", "invalid radix point in binary literal"},
		{token.FLOAT, "0b.1", "0b.1", "invalid radix point in binary literal"},
		{token.FLOAT, "0b1.0", "0b1.0", "invalid radix point in binary literal"},
		{token.FLOAT, "0b1e10", "0b1e10", "'e' exponent requires decimal mantissa"},
		{token.FLOAT, "0b1P-1", "0b1P-1", "'P' exponent requires hexadecimal mantissa"},

		// octals
		{token.INT, "0o0", "0o0", ""},
		{token.INT, "0o1234", "0o1234", ""},
		{token.INT, "0O1234", "0O1234", ""},

		{token.INT, "0o", "0o", "octal literal has no digits"},
		{token.INT, "0o8123", "0o8123", "invalid digit '8' in octal literal"},
		{token.INT, "0o1293", "0o1293", "invalid digit '9' in octal literal"},
		{token.INT, "0o12a3", "0o12 a3", ""}, // only accept 0-9

		{token.FLOAT, "0o.", "0o.", "invalid radix point in octal literal"},
		{token.FLOAT, "0o.2", "0o.2", "invalid radix point in octal literal"},
		{token.FLOAT, "0o1.2", "0o1.2", "invalid radix point in octal literal"},
		{token.FLOAT, "0o1E+2", "0o1E+2", "'E' exponent requires decimal mantissa"},
		{token.FLOAT, "0o1p10", "0o1p10", "'p' exponent requires hexadecimal mantissa"},

		// 0-octals
		{token.INT, "0", "0", ""},
		{token.INT, "0123", "0123", ""},

		{token.INT, "08123", "08123", "invalid digit '8' in octal literal"},
		{token.INT, "01293", "01293", "invalid digit '9' in octal literal"},
		{token.INT, "0F.", "0 F .", ""}, // only accept 0-9
		{token.INT, "0123F.", "0123 F .", ""},
		{token.INT, "0123456x", "0123456 x", ""},

		// decimals
		{token.INT, "1", "1", ""},
		{token.INT, "1234", "1234", ""},

		{token.INT, "1f", "1 f", ""}, // only accept 0-9

		// decimal floats
		{token.FLOAT, "0.", "0.", ""},
		{token.FLOAT, "123.", "123.", ""},
		{token.FLOAT, "0123.", "0123.", ""},

		{token.FLOAT, ".0", ".0", ""},
		{token.FLOAT, ".123", ".123", ""},
		{token.FLOAT, ".0123", ".0123", ""},

		{token.FLOAT, "0.0", "0.0", ""},
		{token.FLOAT, "123.123", "123.123", ""},
		{token.FLOAT, "0123.0123", "0123.0123", ""},

		{token.FLOAT, "0e0", "0e0", ""},
		{token.FLOAT, "123e+0", "123e+0", ""},
		{token.FLOAT, "0123E-1", "0123E-1", ""},

		{token.FLOAT, "0.e+1", "0.e+1", ""},
		{token.FLOAT, "123.E-10", "123.E-10", ""},
		{token.FLOAT, "0123.e123", "0123.e123", ""},

		{token.FLOAT, ".0e-1", ".0e-1", ""},
		{token.FLOAT, ".123E+10", ".123E+10", ""},
		{token.FLOAT, ".0123E123", ".0123E123", ""},

		{token.FLOAT, "0.0e1", "0.0e1", ""},
		{token.FLOAT, "123.123E-10", "123.123E-10", ""},
		{token.FLOAT, "0123.0123e+456", "0123.0123e+456", ""},

		{token.FLOAT, "0e", "0e", "exponent has no digits"},
		{token.FLOAT, "0E+", "0E+", "exponent has no digits"},
		{token.FLOAT, "1e+f", "1e+ f", "exponent has no digits"},
		{token.FLOAT, "0p0", "0p0", "'p' exponent requires hexadecimal mantissa"},
		{token.FLOAT, "1.0P-1", "1.0P-1", "'P' exponent requires hexadecimal mantissa"},

		// hexadecimals
		{token.INT, "0x0", "0x0", ""},
		{token.INT, "0x1234", "0x1234", ""},
		{token.INT, "0xcafef00d", "0xcafef00d", ""},
		{token.INT, "0XCAFEF00D", "0XCAFEF00D", ""},

		{token.INT, "0x", "0x", "hexadecimal literal has no digits"},
		{token.INT, "0x1g", "0x1 g", ""},

		// hexadecimal floats
		{token.FLOAT, "0x0p0", "0x0p0", ""},
		{token.FLOAT, "0x12efp-123", "0x12efp-123", ""},
		{token.FLOAT, "0xABCD.p+0", "0xABCD.p+0", ""},
		{token.FLOAT, "0x.0189P-0", "0x.0189P-0", ""},
		{token.FLOAT, "0x1.ffffp+1023", "0x1.ffffp+1023", ""},

		{token.FLOAT, "0x.", "0x.", "hexadecimal literal has no digits"},
		{token.FLOAT, "0x0.", "0x0.", "hexadecimal mantissa requires a 'p' exponent"},
		{token.FLOAT, "0x.0", "0x.0", "hexadecimal mantissa requires a 'p' exponent"},
		{token.FLOAT, "0x1.1", "0x1.1", "hexadecimal mantissa requires a 'p' exponent"},
		{token.FLOAT, "0x1.1e0", "0x1.1e0", "hexadecimal mantissa requires a 'p' exponent"},
		{token.FLOAT, "0x1.2gp1a", "0x1.2 gp1a", "hexadecimal mantissa requires a 'p' exponent"},
		{token.FLOAT, "0x0p", "0x0p", "exponent has no digits"},
		{token.FLOAT, "0xeP-", "0xeP-", "exponent has no digits"},
		{token.FLOAT, "0x1234PAB", "0x1234P AB", "exponent has no digits"},
		{token.FLOAT, "0x1.2p1a", "0x1.2p1 a", ""},

		// separators
		{token.INT, "0b_1000_0001", "0b_1000_0001", ""},
		{token.INT, "0o_600", "0o_600", ""},
		{token.INT, "0_466", "0_466", ""},
		{token.INT, "1_000", "1_000", ""},
		{token.FLOAT, "1_000.000_1", "1_000.000_1", ""},
		{token.INT, "0x_f00d", "0x_f00d", ""},
		{token.FLOAT, "0x_f00d.0p1_2", "0x_f00d.0p1_2", ""},

		{token.INT, "0b__1000", "0b__1000", "'_' must separate successive digits"},
		{token.INT, "0o60___0", "0o60___0", "'_' must separate successive digits"},
		{token.INT, "0466_", "0466_", "'_' must separate successive digits"},
		{token.FLOAT, "1_.", "1_.", "'_' must separate successive digits"},
		{token.FLOAT, "0._1", "0._1", "'_' must separate successive digits"},
		{token.FLOAT, "2.7_e0", "2.7_e0", "'_' must separate successive digits"},
		{token.INT, "0x___0", "0x___0", "'_' must separate successive digits"},
		{token.FLOAT, "0x1.0_p0", "0x1.0_p0", "'_' must separate successive digits"},
	} {
		var err string
		s := New(fset.AddFile("", fset.Base(), len(test.src)), []byte(test.src), func(_ token.Position, msg string) {
			if err == "" {
				err = msg
			}
		})
		for i, want := range strings.Split(test.tokens, " ") {
			err = ""
			s.state = stateCodeBlock
			tok, lit := s.scanCodeBlock()

			// compute lit where for tokens where lit is not defined
			switch tok {
			case token.DOT:
				lit = "."
			case token.ADD:
				lit = "+"
			case token.SUB:
				lit = "-"
			}

			if i == 0 {
				if tok != test.tok {
					t.Errorf("%q: got token %s; want %s", test.src, tok, test.tok)
				}
				if err != test.err {
					t.Errorf("%q: got error %q; want %q", test.src, err, test.err)
				}
			}

			if lit != want {
				t.Errorf("%q: got literal %q (%s); want %s", test.src, lit, tok, want)
			}
		}

		// make sure we read all
		_, tok, _ := s.Scan()
		if tok != token.EOF {
			t.Errorf("%q: got %s; want EOF", test.src, tok)
		}
	}
}
