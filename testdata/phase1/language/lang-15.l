Use IO's Print from `io` as "Print".
"firstMessage" is `first`.
"secondMessage" is `second`.

    Create "First" as a New Public Basic App.
The App Prints firstMessage.
Show `first-continuation|`.
Exit.
Show `unreachable-local|`.

    Create "Second" as a New Public Basic App.
The App Prints secondMessage.
Show `unreachable-second|`.

Run First.
Run Second.
Show `launched|`.
Await First.
Await Second.
Show `unreachable-root|`.
