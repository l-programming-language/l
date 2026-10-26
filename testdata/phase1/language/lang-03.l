Use IO's Print from `io` as "Print".
"message" is `worker`.
"controllerMessage" is `controller`.
Create "Worker" as a New Basic App.
The App Prints message.

    Create "Controller" as a New Public Basic App.
The App Prints controllerMessage.
Controller awaits Worker.
Show `controller-tail|`.

Run Worker.
Run Controller.
Await Controller.
