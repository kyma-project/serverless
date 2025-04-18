const { createClient } = require('redis');

module.exports = {
    main: async function (event, _) {

        const port = process.env["REDIS_PORT"];
        const host = process.env["REDIS_HOST"];
        const password = process.env["REDIS_PASSWORD"];
   
        const client = await createClient({
            password,
            socket: {
                host,
                port,
            }
        }).on('error', err => 
            console.log('Redis Client Error', err)
        ).connect();

        let result = undefined
        try {
            const req = event.extensions.request
            let key = req.path

            if (key){
                key = key.replace(/^\//g, '');
                if (req.method == "GET") {
                    value = await client.get(key);
                    if(!value){
                        event.extensions.response.status(404); 
                    }
                    result = JSON.parse(value);
                } else if (req.method == "POST") {
                    let value = await client.get(key);
                    if (!value) {
                        value = JSON.stringify(req.body)
                        await client.set(key, value)
                        event.extensions.response.status(201);    
                    } else {
                        event.extensions.response.status(409);    
                    }
                } else if (req.method == "DELETE") {
                    await client.del(key)
                    event.extensions.response.status(204);    
                } else {
                    event.extensions.response.status(405);    
                }
            } else {
                event.extensions.response.status(400);    
            }
        } catch(e) {
            console.error("error ", e)
            result = e    
            event.extensions.response.status(500);    
        } finally {
            await client.disconnect();
            return result;
        }
    }
}


