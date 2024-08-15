# The Final Stockbot Turned into a Congress Tracker

https://www.dirtycongress.com/

Been building various stock trackers trying to take advantage of publically
available information for years. Gave up on that and built something to track
congress. It's shocking how much information is publically available.

# The Code Layout

May or may not be useful to whoever is looking at it.

The app lives in `/internal/app` with various modules living in `/internal` that plug into the overall app.

It attempts to be modular but really isn't that modular.

# Running It

You'll need congress.sqlite in the current directory. Download it from the server.

    DEBUG=true go run . -disable-fetcher

This is the best way to run it locally.