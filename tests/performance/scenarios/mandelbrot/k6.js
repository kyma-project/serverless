import http from "k6/http";
import { sleep } from 'k6';
import exec from 'k6/execution';

export const options = {
    scenarios: {
    warmup: {
        executor: 'constant-vus',
        vus: 5,
        duration: '1m',
        gracefulStop: '0s',
        exec: 'call_function_then_sleep',
    },
    constant_load: {
        executor: 'constant-vus',
        vus: 50,
        startTime: '1m',
        duration: "2m",
        gracefulStop: '0s',
        exec: 'call_function_then_sleep',
    },
    max_load: {
        executor: 'constant-vus',
        vus: 100,
        startTime: '3m',
        duration: "2m",
        gracefulStop: '0s',
        exec: 'call_function',
    },
    ramping_max_load: {
        executor: 'ramping-vus',
        startVUs: 0,
        stages: [
        { duration: '30s', target: 175 },
        { duration: '30s', target: 350 },
        { duration: '30s', target: 525 },
        { duration: '30s', target: 700 },
        { duration: '30s', target: 875 },
        { duration: '30s', target: 1050 },
        ],
        gracefulRampDown: '0s',
        startTime: '5m',
        gracefulStop: '0s',
        exec: 'call_function',
    },
    },
};

export function call_function() {
    get();
};

export function call_function_then_sleep() {
    get();
    sleep(1);
};

function get() {
    http.get(`http://${__ENV.FUNC_HOSTNAME}:80/`, {
        headers: { 
            'Content-Type': 'application/json',
            'width': 800,
            'height': 640,
            'iterations': exec.instance.vusActive,
            'zoom': 1,
        },
    });
};