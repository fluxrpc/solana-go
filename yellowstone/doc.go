// Package yellowstone is a client for the Yellowstone gRPC Geyser plugin
// protocol (Dragon's Mouth), shipped as a nested Go module so its
// gRPC/protobuf dependency tree never reaches consumers of the core SDK:
//
//	go get github.com/fluxrpc/solana-go/yellowstone
//
// # Connecting
//
// [Connect] accepts host:port or URL endpoints; TLS is inferred for
// https:// or :443 targets. Authentication uses x-token metadata on every
// call, and the receive limit defaults to 1GB because geyser block updates
// dwarf gRPC's 4MB default:
//
//	client, err := yellowstone.Connect(ctx, "https://your-geyser:443",
//		yellowstone.WithToken("..."))
//
// # Subscribing
//
// [Client.Subscribe] opens the bidirectional update stream. [Request] owns
// its named filters and exposes fluent builders; [Stream.Update] changes
// filters on a live stream:
//
//	req := yellowstone.NewRequest(pb.CommitmentLevel_CONFIRMED).
//		AccountsByOwner("usdc", tokenProgram).
//		AllSlots("slots")
//	stream, err := client.Subscribe(ctx, req)
//	for {
//		update, err := stream.Recv()
//		...
//	}
//
// # Converters
//
// [Update.Transaction] and [Update.Account] map geyser protobuf payloads into
// the core SDK's types. Converted transactions re-serialize byte-identical to
// the on-chain wire form. For throughput,
// [AccountUpdate].Data aliases the protobuf buffer — copy it if it must
// outlive the next Recv.
package yellowstone
