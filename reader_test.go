package netconf

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"testing"
	"unicode"
)

const (
	SRX240IfaceFileName = "testdata/srx240_interface_info.xml"
)

func TestReader_Read_IOCopy(t *testing.T) {

	srx240IfaceFile, err := os.Open(SRX240IfaceFileName)
	if err != nil {
		t.Fatalf("failed to open test file %q: %v", SRX240IfaceFileName, err)
	}

	defer srx240IfaceFile.Close()

	ncReader := NewReader(srx240IfaceFile)
	var buf bytes.Buffer
	n, err := io.Copy(&buf, ncReader)
	if err != nil {
		t.Errorf("io.Copy error after %d bytes written: %v", n, err)
	}

	// bytes read should be identical to untouched bytes,
	// minus message separator and trailing newline
	srx240IfaceFileBytes, err := ioutil.ReadFile(SRX240IfaceFileName)
	if err != nil {
		t.Fatalf("failed to read all bytes from test file %q: %v", SRX240IfaceFileName, err)
	}

	want := bytes.TrimSuffix(
		bytes.TrimRightFunc(srx240IfaceFileBytes, unicode.IsSpace),
		[]byte(MessageSeparator),
	)

	if !bytes.Equal(want, buf.Bytes()) {
		t.Errorf("Unexpected bytes from Reader using io.Copy:want:\t%q\ngot:\t%q", want, buf.Bytes())
	} else {
		t.Logf("Read correctly read bytes when called by io.Copy:\nwant == got == %q", buf.Bytes())
	}
}

// func TestReader_Read_Bufio_ReadFrom(t *testing.T) { }
