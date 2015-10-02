Notification templates? ok
==========================

You can edit templates in the `src` directory, then run `make` and they will be outputted as go packages, which will be available at `github.com/opsee/notification-templates/dist/go/$type`.

The go package and variable name is built from the file path, so src/slack/new-customer.json.mustache becomes:

```
import "github.com/opsee/notification-templates/dist/go/slack"

slack.NewCustomer //=> the template...
```
