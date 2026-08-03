module.exports = {
  testEnvironment: "node",
  // Redirect require("./mod.js") to our adapter
  moduleNameMapper: {
    "^./mod.js$": "<rootDir>/adapter.js",
  },
};