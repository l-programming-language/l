Use IO's Print from `io` as "Print".
"firstMessage" is `worker`.
"secondMessage" is `other`.

    Create "Other" as a New Public Basic App.
The App Prints secondMessage.
Show `other-continuation|`.

    Create "Worker" as a New Public Basic App.
The App Prints firstMessage.
Run Other.
Show `worker-before-join|`.
Await Other.
Show `worker-after-join|`.

Run Worker.
Show `outer-before-join|`.
Await Worker.
Show `outer-after-join|`.
