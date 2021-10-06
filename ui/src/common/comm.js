export async function easyFetch(url) {
    try {
        let response = await fetch(url);
        if ( response.status !== 200 ) {
            return ['error', [
                'server responded with status: ' + response.status,
                { url: url },
                await response.text(),
            ]];
        }
        let data = await response.json();
        return [false, data];
    } catch (e) {
        return ['error', [
            'an exception occurred during a request',
            { url: url },
            e.message
        ]];
    }
}

export async function easyPost(url, payload) {
    console.log('payload', JSON.stringify(payload))
    try {
        let response = await fetch(url, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });
        if ( response.status !== 200 ) {
            return ['error', [
                'server responded with status: ' + response.status,
                { url: url },
                await response.text(),
            ]];
        }
        let data = await response.json();
        return [false, data]
    } catch (e) {
        return ['error', [
            'an exception occurred during a request',
            { url: url },
            e.message
        ]]
    }
}