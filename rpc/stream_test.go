package rpc

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/bytedance/sonic"
)

func streamAccountJSON(i int) string {
	return fmt.Sprintf(`{"pubkey":"SysvarC1ock11111111111111111111111111111111","account":{"lamports":%d,"owner":"11111111111111111111111111111111","data":["dGVzdCBkYXRh","base64"],"executable":false,"rentEpoch":361,"space":9}}`, i+1)
}

func streamEnvelope(withContext bool, n int) string {
	var accounts []string
	for i := 0; i < n; i++ {
		accounts = append(accounts, streamAccountJSON(i))
	}
	list := "[" + strings.Join(accounts, ",") + "]"
	if withContext {
		return `{"jsonrpc":"2.0","result":{"context":{"slot":341197053,"apiVersion":"2.0.15"},"value":` + list + `},"id":1}`
	}
	return `{"jsonrpc":"2.0","result":` + list + `,"id":1}`
}

func TestStreamProgramAccounts(t *testing.T) {
	for _, withContext := range []bool{false, true} {
		t.Run(fmt.Sprintf("withContext=%v", withContext), func(t *testing.T) {
			var got []*KeyedAccount
			ctx, err := StreamProgramAccounts(strings.NewReader(streamEnvelope(withContext, 3)), func(ka *KeyedAccount) error {
				got = append(got, ka)
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != 3 {
				t.Fatalf("got %d accounts", len(got))
			}
			if got[1].Account.Lamports != 2 || string(got[1].Account.Data.GetBinary()) != "test data" {
				t.Fatalf("account[1] = %+v", got[1].Account)
			}
			if withContext {
				if ctx == nil || ctx.Slot != 341197053 {
					t.Fatalf("context = %+v", ctx)
				}
			} else if ctx != nil {
				t.Fatalf("unexpected context %+v", ctx)
			}
		})
	}
}

func TestStreamProgramAccountsEdgeCases(t *testing.T) {
	// Null result: no accounts, no error.
	count := 0
	if _, err := StreamProgramAccounts(strings.NewReader(`{"jsonrpc":"2.0","result":null,"id":1}`), func(*KeyedAccount) error {
		count++
		return nil
	}); err != nil || count != 0 {
		t.Fatalf("null result: count %d, err %v", count, err)
	}

	// JSON-RPC error envelope surfaces as *RPCError.
	var rpcErr *RPCError
	_, err := StreamProgramAccounts(strings.NewReader(`{"jsonrpc":"2.0","error":{"code":-32601,"message":"nope"},"id":1}`), func(*KeyedAccount) error { return nil })
	if !errors.As(err, &rpcErr) || rpcErr.Code != -32601 {
		t.Fatalf("err = %v", err)
	}

	// Callback errors abort the stream.
	boom := errors.New("boom")
	calls := 0
	_, err = StreamProgramAccounts(strings.NewReader(streamEnvelope(false, 3)), func(*KeyedAccount) error {
		calls++
		return boom
	})
	if !errors.Is(err, boom) || calls != 1 {
		t.Fatalf("calls %d, err %v", calls, err)
	}

	// Malformed body errors instead of hanging.
	if _, err := StreamProgramAccounts(strings.NewReader(`[1,2,3]`), func(*KeyedAccount) error { return nil }); err == nil {
		t.Fatal("accepted a non-envelope body")
	}
}

// TestStreamProgramAccountsIsIncremental proves accounts are delivered while
// the response body is still arriving: the first two accounts are decoded
// before the writer has produced the rest of the payload.
func TestStreamProgramAccountsIsIncremental(t *testing.T) {
	envelope := streamEnvelope(true, 4)
	// Split just after the 2nd account object ends: the decoder can finish
	// two elements without ever seeing the rest.
	second := streamAccountJSON(1)
	cut := strings.Index(envelope, second) + len(second)

	pr, pw := io.Pipe()
	delivered := make(chan int)
	done := make(chan error, 1)

	go func() {
		count := 0
		_, err := StreamProgramAccounts(pr, func(*KeyedAccount) error {
			count++
			delivered <- count
			return nil
		})
		done <- err
	}()

	if _, err := pw.Write([]byte(envelope[:cut])); err != nil {
		t.Fatal(err)
	}
	// Both fully-transmitted accounts must arrive with the tail unwritten.
	if got := <-delivered; got != 1 {
		t.Fatalf("first delivery = %d", got)
	}
	if got := <-delivered; got != 2 {
		t.Fatalf("second delivery = %d", got)
	}

	if _, err := pw.Write([]byte(envelope[cut:])); err != nil {
		t.Fatal(err)
	}
	pw.Close()
	if got := <-delivered; got != 3 {
		t.Fatalf("third delivery = %d", got)
	}
	if got := <-delivered; got != 4 {
		t.Fatalf("fourth delivery = %d", got)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
}

// pacedReader delivers its data on a fixed schedule — chunk i becomes
// readable at start + i*interval — modeling a network stream that keeps
// arriving in the background while the consumer is busy decoding. Reads
// return whatever has "arrived" and block only when the consumer outruns
// the link.
type pacedReader struct {
	data     []byte
	pos      int
	chunk    int
	interval time.Duration
	start    time.Time
}

func (r *pacedReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	if r.start.IsZero() {
		r.start = time.Now()
	}
	for {
		elapsed := time.Since(r.start)
		arrived := int(elapsed/r.interval+1) * r.chunk
		if arrived > len(r.data) {
			arrived = len(r.data)
		}
		if r.pos < arrived {
			n := copy(p, r.data[r.pos:arrived])
			r.pos += n
			return n, nil
		}
		next := r.start.Add(time.Duration(r.pos/r.chunk+1) * r.interval)
		time.Sleep(time.Until(next))
	}
}

// The paced pair demonstrates the streaming win: with data arriving over a
// link, buffered decode pays download + decode sequentially, while the
// stream decodes during the transfer and finishes with the download. The
// ttfa-ns metric is the time until the first account is delivered.
func benchmarkPacedEnvelope() []byte { return []byte(streamEnvelope(true, 2000)) }

func BenchmarkStreamProgramAccountsNetwork(b *testing.B) {
	envelope := benchmarkPacedEnvelope()
	b.ReportAllocs()
	var firstTotal time.Duration
	for b.Loop() {
		r := &pacedReader{data: envelope, chunk: 32 << 10, interval: 250 * time.Microsecond}
		started := time.Now()
		first := time.Duration(0)
		count := 0
		if _, err := StreamProgramAccounts(r, func(*KeyedAccount) error {
			if count == 0 {
				first = time.Since(started)
			}
			count++
			return nil
		}); err != nil || count != 2000 {
			b.Fatal(err, count)
		}
		firstTotal += first
	}
	b.ReportMetric(float64(firstTotal.Nanoseconds())/float64(b.N), "ttfa-ns/op")
}

func BenchmarkBufferedProgramAccountsNetwork(b *testing.B) {
	envelope := benchmarkPacedEnvelope()
	b.ReportAllocs()
	var firstTotal time.Duration
	for b.Loop() {
		r := &pacedReader{data: envelope, chunk: 32 << 10, interval: 250 * time.Microsecond}
		started := time.Now()
		body, err := io.ReadAll(r)
		if err != nil {
			b.Fatal(err)
		}
		var out struct {
			Result GetProgramAccountsWithContextResult `json:"result"`
		}
		if err := sonic.Unmarshal(body, &out); err != nil {
			b.Fatal(err)
		}
		if len(out.Result.Value) != 2000 {
			b.Fatal("bad count")
		}
		firstTotal += time.Since(started) // first account usable only now
	}
	b.ReportMetric(float64(firstTotal.Nanoseconds())/float64(b.N), "ttfa-ns/op")
}

var benchmarkStreamCount int

// The benchmark compares total CPU cost against buffering + whole-slice
// decode. The stream's real advantage — overlapping decode with network
// download and never holding the full response — does not show in ns/op.
func BenchmarkStreamProgramAccounts(b *testing.B) {
	envelope := streamEnvelope(true, 200)
	b.ReportAllocs()
	for b.Loop() {
		benchmarkStreamCount = 0
		_, err := StreamProgramAccounts(strings.NewReader(envelope), func(*KeyedAccount) error {
			benchmarkStreamCount++
			return nil
		})
		if err != nil || benchmarkStreamCount != 200 {
			b.Fatal(err, benchmarkStreamCount)
		}
	}
}

func BenchmarkBufferedProgramAccounts(b *testing.B) {
	envelope := streamEnvelope(true, 200)
	b.ReportAllocs()
	for b.Loop() {
		var out struct {
			Result GetProgramAccountsWithContextResult `json:"result"`
		}
		if err := sonic.Unmarshal([]byte(envelope), &out); err != nil {
			b.Fatal(err)
		}
		if len(out.Result.Value) != 200 {
			b.Fatal("bad count")
		}
	}
}
