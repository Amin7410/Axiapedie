// Mock browser environment
global.window = global;
global.window.addEventListener = () => {};
global.document = {
    addEventListener: () => {},
    getElementById: () => ({
        addEventListener: () => {}
    }),
    querySelectorAll: () => [],
    body: {
        getAttribute: () => "admin"
    }
};
global.localStorage = {
    getItem: () => null,
    setItem: () => {}
};

try {
    console.log("Loading dialogs.js...");
    require("./web/static/js/dialogs.js");
    console.log("Loading sidebar.js...");
    require("./web/static/js/sidebar.js");
    console.log("Loading explorer.js...");
    require("./web/static/js/explorer.js");
    console.log("Loading search.js...");
    require("./web/static/js/search.js");
    console.log("All scripts loaded successfully without runtime errors!");
} catch (err) {
    console.error("Runtime error detected during loading:", err);
}
