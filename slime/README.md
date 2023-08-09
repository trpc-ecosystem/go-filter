
# Slime

Slime is the implementation fo tRPC-Go retry/hedging policy.

A request in application layer may spans multiple sub-requests after retry/hedging interceptor.
At last, The first success response or last failed response will be delivered to application.
Because the procedure look like a slime, splitting and fusing, we call it Slime.

See [iwiki](https://trpc.group/trpc-go/trpc-wiki/blob/main/user_guide/retry_hedging.md) for how to use Slime. 
