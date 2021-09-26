const AWS = require('aws-sdk');
const sns = new AWS.SNS()
const TOPIC_ARN = process.env.TOPIC_ARN;

const headers = {
    'Content-Type': 'application/json',
    'x-slack-no-retry': '1',
};

exports.handler = (event) => {
    if (!event.body) {
        return Promise.resolve({ statusCode: 400, body: 'invalid', headers: headers });
    }
    console.log("EVENT BODY", event.body)
    return sns.publish({
        Message: event.body,
        TopicArn: TOPIC_ARN
    })
        .promise()
        .then(() => ({ statusCode: 200, body: event.body, headers: headers }))
        .catch(err => {
            console.log(err);
            return { statusCode: 500, body: 'sns-error', headers: headers };
        });
};