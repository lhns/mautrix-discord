# mautrix-discord
A Matrix-Discord puppeting bridge based on [discordgo](https://github.com/bwmarrin/discordgo).

## Documentation
All setup and usage instructions are located on [docs.mau.fi]. Some quick links:

[docs.mau.fi]: https://docs.mau.fi/bridges/go/discord/index.html

* [Bridge setup](https://docs.mau.fi/bridges/go/setup.html?bridge=discord)
  (or [with Docker](https://docs.mau.fi/bridges/general/docker-setup.html?bridge=discord))
* Basic usage: [Authentication](https://docs.mau.fi/bridges/go/discord/authentication.html),
  [Relaying with webhooks](https://docs.mau.fi/bridges/go/discord/relay.html)

### Features & Roadmap
[ROADMAP.md](https://github.com/mautrix/discord/blob/main/ROADMAP.md)
contains a general overview of what is supported by the bridge.

## Discussion
Matrix room: [#discord:maunium.net](https://matrix.to/#/#discord:maunium.net)

## Fork patches (lhns)
Built from the `fork` branch, see `.github/workflows/build-image.yml`.

* Backfill: configurable fetch delay and bounded retry.
* DM relay (incl. group DMs) through the receiver's own account.
* Untitled group DMs are named after their participants (`{{.Participants}}` in `channel_name_template`).
* Exit the process when a Discord message can't be sent to Matrix because the homeserver is unavailable (network error, 429, 502-504), so the restart's missed-message backfill recovers it instead of it being dropped. Needs `bridge.backfill.forward_limits.missed` non-zero for the channel type, and a previously bridged message in the portal.
