# Optional: a meeting point for the phone app

Without anything set up, the phone link is a random Cloudflare address that
changes each game session (scan the QR code or type the pairing code each
time). A meeting point gives the phone app **one fixed address, for good**,
and the app then finds your Mac by itself every session:

- Open the app from the Home Screen, before or after starting the game: it
  says "Waiting for the game", then connects by itself.
- Several Macs with the same phone: Settings > **Computers** lists them; the
  app connects to whichever has the game running.
- You only need a **free Cloudflare account**, no domain.

The meeting point is a small Cloudflare Worker that **you** deploy to your own
account (nobody else hosts anything for you). It only keeps, for each Mac:
an ID, its name, today's tunnel address, and while a pairing code is shown,
that code hashed (5 minutes). No game data goes through it, and never the
pairing key. Without it, or if it is down, everything works as before.

## Set up (once)

You need Node.js on the Mac (https://nodejs.org) and the SimSight source
folder (this repository), built once with `mac/build.sh`.

1. Deploy the Worker (it opens the browser once to log in to Cloudflare):

   ```sh
   cd companion/meeting-point
   npx wrangler login
   npx wrangler deploy
   ```

   It prints the address, like `https://simsight.<your-account>.workers.dev`.
2. On the Mac, open
   `~/Library/Application Support/SimSightCompanion/config.json` and add one
   line (keep the others):

   ```json
   "meeting_point": "https://simsight.<your-account>.workers.dev"
   ```
3. Start the game and load a household. Option+Command+A, **phone
   companion**: the QR code now opens the app at the meeting point, and the
   Mac says a 6-digit code.
4. On the phone: open the meeting point address, **Share > Add to Home
   Screen**, open the app from its icon and type the code (or scan the QR
   code in Safari first, then add it).

From then on: just open the app.

**Another Mac:** set the same `meeting_point` in that Mac's config.json; on
the phone, Settings > Computers > **Pair another computer**, type the code
that Mac says.

## Good to know

- **Update the Worker when you update the mod:** the phone app lives there,
  so run `npx wrangler deploy` again after `mac/build.sh`. An old app with a
  newer mod may miss buttons.
- **Your own domain** (optional): in the Cloudflare dashboard, Workers >
  simsight > Settings > Domains & Routes > add a custom domain such as
  `sims-app.example.com`. Use a different name from a fixed tunnel address
  ([FIXED-ADDRESS.md](FIXED-ADDRESS.md)) if you have one; the two do not
  need each other.
- **Free plan:** plenty for a household (a few thousand requests a day).
- **Turn it off:** remove the `meeting_point` line. Phones then use the
  Mac's own link again (pair them again once).
- **Forget a phone or Mac:** on the phone, Settings > Computers > Forget.
