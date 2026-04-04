const { test } = require('node:test');
const assert = require('node:assert');

const getTableData = (data, smartTable) => {
    if (!smartTable) return null;
    const content = data;

    if (Array.isArray(content) && content.length > 0) {
        const isListOfObjects = content.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
        if (isListOfObjects) {
            return content;
        }
        const isListOfPrimitives = content.every(item => typeof item !== 'object' || item === null);
        if (isListOfPrimitives) {
            return content.map((item, index) => ({ index, value: item }));
        }
    }

    if (content && typeof content === 'object' && !Array.isArray(content) && content !== null) {
        let largestArray = null;
        Object.values(content).forEach(val => {
            if (Array.isArray(val) && val.length > 0) {
                const isListOfObjects = val.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
                if (isListOfObjects) {
                    if (!largestArray || val.length > largestArray.length) {
                        largestArray = val;
                    }
                } else {
                    const isListOfPrimitives = val.every(item => typeof item !== 'object' || item === null);
                    if (isListOfPrimitives) {
                        const mapped = val.map((item, index) => ({ index, value: item }));
                        if (!largestArray || mapped.length > largestArray.length) {
                            largestArray = mapped;
                        }
                    } else {
                        const mapped = val.map((item, index) => ({ index, value: typeof item === 'object' ? JSON.stringify(item) : item }));
                        if (!largestArray || mapped.length > largestArray.length) {
                            largestArray = mapped;
                        }
                    }
                }
            } else if (val && typeof val === 'object' && !Array.isArray(val) && val !== null) {
                Object.values(val).forEach(nestedVal => {
                    if (Array.isArray(nestedVal) && nestedVal.length > 0) {
                        const isListOfObjects = nestedVal.every(item => typeof item === 'object' && item !== null && !Array.isArray(item));
                        if (isListOfObjects) {
                            if (!largestArray || nestedVal.length > largestArray.length) {
                                largestArray = nestedVal;
                            }
                        } else {
                            const isListOfPrimitives = nestedVal.every(item => typeof item !== 'object' || item === null);
                            if (isListOfPrimitives) {
                                const mapped = nestedVal.map((item, index) => ({ index, value: item }));
                                if (!largestArray || mapped.length > largestArray.length) {
                                    largestArray = mapped;
                                }
                            } else {
                                const mapped = nestedVal.map((item, index) => ({ index, value: typeof item === 'object' ? JSON.stringify(item) : item }));
                                if (!largestArray || mapped.length > largestArray.length) {
                                    largestArray = mapped;
                                }
                            }
                        }
                    }
                });
            }
        });
        if (largestArray) {
            return largestArray;
        }
    }

    return null;
};

const d1 = { "results": ["report_q3.pdf", "data_q3.xlsx"] };
console.log(getTableData(d1, true));
