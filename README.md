*UNDER CONSTRUCTION* This is a fork of the amazing [xanderstrike/goplaxt](https://github.com/xanderstrike/goplaxt) to translate its functionality to [AniList](https://anilist.co/).

# AniPlaxt

Plex provides webhook integration for all Plex Pass subscribers, and users of their servers. A webhook is a request that the Plex application sends to third party services when a user takes an action, such as watching a movie or episode.

You can ask Plex to send these webhooks to this tool, which will then log those plays in your AniList account.

To start scrobbling today, head to [aniplaxt.natwelch.com](https://aniplaxt.natwelch.com) and enter your Plex username!

It's as easy as can be!

If you experience any problems or have any suggestions, please don't hesitate to create an issue on this repo.

### Deploying For Yourself

TODO: We don't currently support this. If you'd like to, feel free, as this code is MIT licensed. I'd love a PR with updated description on how to.

### Contributing

Please do! I accept any and all PRs.

### Security PSA

You should know that by using the instance I host, I perminantly retain your
Plex username, and an API key that allows me to add plays to your AniList
account (but not your username). Also, I log the title and year of films you
watch and the title, season, and episode of shows you watch. These logs are
temporary and are rotated every 24 hours with older logs perminantly deleted.

I promise to Do No Harm with this information. It will never leave my server,
and I won't look at it unless I'm troubleshooting bugs. Frankly, I couldn't
care less. However, I believe it's important to disclose my access to your
information. If you are not comfortable sharing I encourage you to host the
application on your own hardware.

[I have never been served with any government requests for data](https://en.wikipedia.org/wiki/Warrant_canary).
