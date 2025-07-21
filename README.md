# vxFormsUI

GO service for processing ingest related forms
This project is a Go module that uses the [Gin web framework](https://gin-gonic.com/) and [Bootstrap 5](https://getbootstrap.com/) to implement a web-based UI with the following features:

- **Form Selection:** Users can choose which form to use from a main page.
- **Dynamic Form Rendering:** The UI presents the appropriate form for creating the associated JSON document based on user input.
- **Default Version Value:** Any input field named `version` is pre-filled with the default value `"V01"`.
- **Back Navigation:** Each form includes a back button that returns the user to the main page.

