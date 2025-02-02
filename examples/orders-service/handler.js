const redis = require("redis");
const { promisify } = require("util");

let storage = undefined;
const errors = {
    codeRequired: new Error("orderCode is required"),
    alreadyExists: new Error("object already exists"),
}

module.exports = {
    main: async function (event, _) {
        const storage = getStorage();

        if (!event.data || !Object.keys(event.data).length) {
            return await onList(storage, event);
        }

        const { orderCode, consignmentCode, consignmentStatus } = event.data;
        if (orderCode && consignmentCode && consignmentStatus) {
            return await onCreate(storage, event);
        }

        event.extensions.response.status(500);
    }
}

async function onList(storage, event) {
    try {
        return await storage.getAll();
    } catch(err) {
        event.extensions.response.status(500);
        return;
    }
}

async function onCreate(storage, event) {
    try {
        await storage.set(event.data);
    } catch(err) {
        let status = 500;
        switch (err) {
            case errors.codeRequired: {
                status = 400;
                break;
            };
            case errors.alreadyExists: {
                status = 409;
                break;
            };
        }
        event.extensions.response.status(status);
    }
}

class RedisStorage {
    storage = undefined;
    asyncGet = void 0;
    asyncKeys = void 0;
    asyncSet = void 0;

    constructor(options) {
        this.storage = redis.createClient(options);
        this.asyncGet = promisify(this.storage.get).bind(this.storage);
        this.asyncKeys = promisify(this.storage.keys).bind(this.storage);
        this.asyncSet = promisify(this.storage.set).bind(this.storage);
    }

    async getAll() {
        let values = [];

        const keys = await this.asyncKeys("*");
        for (const key of keys) {
            const value = await this.asyncGet(key);
            values.push(JSON.parse(value));
        }

        return values;
    }

    async set(order = {}) {
        if (!order.orderCode) {
            throw errors.codeRequired;
        }
        const value = await this.asyncGet(order.orderCode);
        if (value) {
            throw errors.alreadyExists;
        }
        await this.asyncSet(order.orderCode, JSON.stringify(order));
    }
}

class InMemoryStorage {
    storage = new Map();

    getAll() {
        return Array.from(this.storage)
            .map(([_, order]) => order)
    }

    set(order = {}) {
        if (!order.orderCode) {
            throw errors.codeRequired;
        }
        if (this.storage.get(order.orderCode)) {
            throw errors.alreadyExists;
        }
        return this.storage.set(order.orderCode, order);
    }
}

function readEnv(env = "") {
    return process.env[env] || undefined;
}

function createStorage() {
    let redisPrefix = readEnv("APP_REDIS_PREFIX");
    if (!redisPrefix) {
        redisPrefix = "REDIS_";
    }
    const port = readEnv(redisPrefix + "PORT");
    const host = readEnv(redisPrefix + "HOST");
    const password = readEnv(redisPrefix + "REDIS_PASSWORD");

    if (host && port && password) {
        return new RedisStorage({ host, port, password });
    }
    return new InMemoryStorage();
}

function getStorage() {
    if (!storage) {
        storage = createStorage();
    }
    return storage;
}