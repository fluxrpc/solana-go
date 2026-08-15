// Package ws is a Solana WebSocket pubsub client covering all nine
// subscriptions (account, program, logs, signature, slot, slotsUpdates,
// root, vote, block), built on gobwas/ws for low-level frame control.
//
// # Model
//
// [Connect] dials the endpoint and starts a single read loop that reuses
// one message buffer, answers control frames inline, and routes exact-size
// payload copies to buffered per-subscription channels. Each subscribe
// method returns a typed, generic [Subscription]; decoding happens on the
// consumer's goroutine in Recv, so one slow decode never stalls the
// socket:
//
//	client, err := ws.Connect(ctx, "wss://your-endpoint")
//	sub, err := client.AccountSubscribe(ctx, account, rpc.CommitmentConfirmed)
//	for {
//		update, err := sub.Recv(ctx)
//		...
//	}
//
// # Backpressure
//
// When a subscriber falls behind its buffer (Options.SubscriptionBuffer,
// default 256), notifications for that subscription are dropped and
// counted ([Subscription.Dropped]) instead of blocking every other
// subscription. Size the buffer for your burst profile if drops matter.
//
// # Delivery guarantees
//
// Subscription channels are registered inside the read loop's handling of
// the subscribe ack, so notifications arriving immediately behind the ack
// are never lost. When the connection dies, every Recv returns an error
// wrapping [ErrSubscriptionClosed] and [Client.Err] reports the cause; the
// client does not reconnect automatically.
package ws
