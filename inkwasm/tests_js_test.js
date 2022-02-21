(() => {
    globalThis.TestExported = function (e, e2) {
        if (e.id !== e2) {
            return 0
        }
        return e.id
    }
    globalThis.TestAlignment = function (b, v) {
        if (b) {
            return v
        }
        return 0
    }
    globalThis.TestAlignment2 = function (b, v) {
        if (b[0] === b[1] && b[1] === b[2]) {
            return b[0]
        }
        return 0
    }
    globalThis.TestObjectType_Bool = function (e) {
        return e
    }
    globalThis.TestObjectType_String = function (e) {
        return e === "Hello, 世界"
    }
    globalThis.TestEcho = function (e) {
        if (e instanceof Uint8Array) {
            return e.slice(0, e.byteLength)
        }
        return e
    }
    globalThis.TestObjectType_Object = function (e) {
        return e === globalThis.TestObjectType_String
    }
    globalThis.TestObject_Bytes = function (e) {
        return new Uint8Array([0x00, 0x01, 0x02, 0x03])
    }
    globalThis.TestObject_GetRandom = function (e) {
        crypto.getRandomValues(e)
    }
})();