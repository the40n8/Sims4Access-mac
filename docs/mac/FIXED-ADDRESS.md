# Optional: a fixed address for the phone app

By default the phone link uses a random Cloudflare address that changes every
time the game starts, so the QR code must be scanned again each session and a
Home Screen app stops working. With a fixed address, the phone keeps the same
link: add it to the Home Screen once and open it like an app.

You need a free Cloudflare account and a domain managed by Cloudflare. No
domain? A [meeting point](MEETING-POINT.md) gives the app a fixed address
too, with only a free account.

## Set up (once)

1. In the Cloudflare dashboard: **Zero Trust > Networks > Tunnels > Create a
   tunnel**, type **Cloudflared**, give it a name (for example `sims`).
2. On the install page, copy the **token** (the long text after
   `--token` in the commands shown). Do not run those commands: the mod
   runs the tunnel itself while the game is open.
3. **Public hostname**: choose a subdomain, for example `sims.example.com`,
   service **HTTP**, URL `localhost:47800`. Save.
4. On the Mac, open
   `~/Library/Application Support/SimSightCompanion/config.json` and add the
   two lines (keep the existing `token` and `port`):

   ```json
   {
     "token": "…keep…",
     "port": 47800,
     "tunnel_token": "the token from step 2",
     "public_url": "https://sims.example.com"
   }
   ```
5. Start the game. Once a household is loaded the Mac says "Phone link
   ready…"; the QR code (Option+Command+A, phone companion) now shows the
   fixed address. Scan it once, then **Share > Add to Home Screen**. If the
   Home Screen app ever says **Pair this phone**, choose phone companion on
   the Mac and type the 6-digit code it says: no scanning needed.

## Or from the command line

With `cloudflared` logged in to your account (`cloudflared tunnel login`):

```sh
cloudflared tunnel create simsight
cloudflared tunnel route dns simsight sims.example.com
cloudflared tunnel token simsight     # the tunnel_token for config.json
```

No public hostname needed in the dashboard: the mod sends everything the
tunnel receives to the phone app.

## Notes

- The phone still needs the pairing key that the QR code carries: knowing
  the address alone is not enough to see or control the game.
- For an extra lock, Cloudflare Access can ask for an email code before the
  page opens (Zero Trust > Access > Applications).
- To go back to random addresses, remove the two lines.
- `config.json` holds secrets: do not share it.
