## Global Variables

These are the things you can or must define in the Lua code when creating a module:

* `title: string (optional)` - Just a title to be displayed inside the LNbits UI.
* `description: string (optional)` - Same as `title`. This can be Markdown, and it accepts a special variable `$extBase` which points to the external root of the module's endpoints.
* `models: array of Model` - This is where you declare the data that will be used and saved by the module. You can have multiple collections of different kinds of data. And they will be displayed as tables in your module's LNbits internal UI, accessible through the module's REST API and from the rest of the module's code. Each `Model` is comprised of:
  * `name: string` - The internal name that identifies the collection. Should probably be called "id".
  * `display: string (optional)` - A name to be displayed in the UI, if not specified `name` will be used.
  * `plural: string (optional)` - Like `display`, but for contexts in which the name should be pluralized.
  * `fields: array of Field` - This is the core of the model definition. Here you specify the fields each item in the collection will have, should have or must have. It's comprised of an array of:
    * `name: string` - The internal name that identifies the field. Same as model's `name`.
    * `display: string (optional)` - Same as model's `display`, but for the field.
    * `type: string` - One of:
      * `"string"` - The field accepts text values.
      * `"number"` - The field accepts number values.
      * `"boolean"` - The field accepts boolean values.
      * `"msatoshi"` - Similar to `number`, but with special features.
      * `"url"` - Similar to `string`, but with special features.
      * `"currency"` - A `select` with a predefined list of currencies.
      * `"ref"` - The name of another model in this module to be cross-linked.
    * `options: array of strings or numbers (optional)` - Specify this when you have chosen `type: "select"`.
    * `as: string (optional)` - Specify this as the name of the field from the model specified in `ref` you want to display.
    * `default: string (optional)` - The default value for this field.
    * `required: boolean (optional)` - If this field is required.
  `single: bool (optional)` - This is for when you want your collection to be actually just a single item. It changes the way models are displayed in the UI (instead of a table, it will be just a form with the fields). Useful for when you want the module user to specify settings, for example.
  * `default_filters: object (optional)` - Collections can be filtered manually through on the LNbits UI. This is a way to specify the default filters for better UX.
  * `default_sort: string (optional)` - Collections can be sortered manually on the LNbits UI. This is a way to specify the default sort criteria for better UX. A sort definition string is the name of the field plus `asc` or `desc`, like `kind desc`. By default items will be sorted by creation date.
* `actions: a map of Action (optional)` - The keys of the map should be the names of the actions. The values are objects comprised of:
  * `fields: array of Field` - Same as Model's `fields`. These will be validated and also shown on the LNbits internal UI.
  * `handler: a function` - A function that takes an argument `params` which contains the passed fields.
* `triggers: a map of functions (optional)` - The keys of the map should be one of the existing triggers, the function is whatever you want to do when that trigger fires. The arguments received by the function depend on the type of trigger.
