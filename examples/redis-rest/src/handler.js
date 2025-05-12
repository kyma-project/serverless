const { createClient } = require('redis');

async function initializeRedisClient() {
    const port = process.env["REDIS_PORT"];
    const host = process.env["REDIS_HOST"];
    const password = process.env["REDIS_PASSWORD"];

    const client = createClient({
        password,
        socket: {
            host,
            port,
        }
});

client.on('error', err => console.log('Redis Client Error', err));
    await client.connect();
    return client;
}

async function handleGetRequest(client, key, response) {
    const value = await client.get(key);
    if (!value) {
        response.status(404);
        return null;
    }
    return JSON.parse(value);
}

async function handlePostRequest(client, key, body, response) {
    let value = await client.get(key);
    if (!value) {
        value = JSON.stringify(body);
        await client.set(key, value);
        response.status(201);
    } else {
        response.status(409);
    }
}

async function handleDeleteRequest(client, key, response) {
    await client.del(key);
    response.status(204);
}

module.exports = {
    main: async function (event, _) {
        const client = await initializeRedisClient();
        let result;

        try {
            const req = event.extensions.request;
            const response = event.extensions.response;
            let key = req.path;
            console.log("Key:", key);
            // Remove leading slash from the key
            // and check if it is not empty
            if (key && key.length > 1) {
                key = key.substring(1);
                switch (req.method) {
                    case "GET":
                        result = await handleGetRequest(client, key, response);
                        break;
                    case "POST":
                        await handlePostRequest(client, key, req.body, response);
                        break;
                    case "DELETE":
                        await handleDeleteRequest(client, key, response);
                        break;
                    default:
                        response.status(405);
                }
            } else {
                response.status(400);
                result = { error: "path parameter is required" };
            }
        } catch (e) {
            console.error("Error:", e);
            result = e;
            event.extensions.response.status(500);
        } finally {
            await client.disconnect();
                return result;
        }
    }
}