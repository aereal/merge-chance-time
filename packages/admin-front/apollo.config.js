const path = require("path")

module.exports = {
  client: {
    addTypename: true,
    includes: ["src/**/*.{tsx,ts}"],
    service: {
      localSchemaFile: path.join(__dirname, "../../schema.gql"),
    },
  },
}
