CREATE TABLE IF NOT EXISTS appSettings (
    id INTEGER PRIMARY KEY,
    maintenance BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER NOT NULL UNIQUE,
    chatModelId INTEGER NOT NULL,
    imageModelId INTEGER NOT NULL,
    tariffId INTEGER NOT NULL,
    lastLimitReset TEXT NOT NULL,
    selfBlock BOOLEAN NOT NULL DEFAULT 0,
    blocked BOOLEAN NOT NULL DEFAULT 0,
    blockReason TEXT,
    FOREIGN KEY (chatModelId) REFERENCES aiModels(id) ON DELETE CASCADE,
    FOREIGN KEY (imageModelId) REFERENCES aiModels(id) ON DELETE CASCADE,
    FOREIGN KEY (tariffId) REFERENCES tariffs(id) ON DELETE CASCADE
);

CREATE TABLE dialogs (
    id INTEGER PRIMARY KEY,
    userId INTEGER NOT NULL,
    title TEXT NOT NULL,
    created TEXT,
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE chatMessages (
    id INTEGER PRIMARY KEY,
    dialogId INTEGER NOT NULL,
    "order" INTEGER NOT NULL,
    "role" INTEGER NOT NULL,
    content TEXT NOT NULL,
    created TEXT,
    FOREIGN KEY (dialogId) REFERENCES dialogs(id) ON DELETE CASCADE
);

CREATE TABLE activeDialogs (
    userId INTEGER PRIMARY KEY,
    dialogId INTEGER NOT NULL,
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (dialogId) REFERENCES dialogs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS aiModels (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    apiName TEXT NOT NULL UNIQUE,
    modelType INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS tariffs (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL UNIQUE,
    rubPrice INTEGER NOT NULL,
    usdPrice INTEGER NOT NULL,
    available BOOLEAN NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tariffLimits (
    id INTEGER PRIMARY KEY,
    tariffId INTEGER NOT NULL,
    aiModelId INTEGER NOT NULL,
    count INTEGER NOT NULL,
    FOREIGN KEY (tariffId) REFERENCES tariffs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS usersUsage (
    userId INTEGER,
    aiModelId INTEGER,
    count INTEGER NOT NULL,
    PRIMARY KEY (userId, aiModelId),
    FOREIGN KEY (userId) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (aiModelId) REFERENCES aiModels(id) ON DELETE CASCADE
);

-- Seeds
INSERT INTO appSettings (id, maintenance) VALUES (1, 0);

INSERT INTO aiModels (id, title, apiName, modelType) VALUES 
(1, 'OpenAI o4-mini', 'o4-mini', 0),
(2, 'OpenAI dall-e-3', 'dall-e-3', 1);

INSERT INTO tariffs (id, Title, rubPrice, usdPrice, available) VALUES 
(1, 'Free', -1, -1, false),
(2, 'PLus', 49900, 500, true),
(3, 'Unlimited', -1, -1, false);

INSERT INTO tariffLimits (aiModelId, tariffId, count) VALUES 
(1, 1, 30),
(1, 2, 100),
(1, 3, -1),
(2, 3, -1);
