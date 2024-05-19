(() => {
    let StringEncoder = new TextEncoder();
    let StringDecoder = new TextDecoder();

    let Objects = [];
    let ObjectsUnused = [];

    let ObjectTypes = {
        TypeUndefined: 0,
        TypeNull: 1,
        TypeBoolean: 2,
        TypeNumber: 3,
        TypeBigInt: 4,
        TypeString: 5,
        TypeSymbol: 6,
        TypeFunction: 7,
        TypeObject: 8,
    }

    if (globalThis.inkwasm === undefined) {
        globalThis.inkwasm = {};
    }
    if (globalThis.inkwasm.Load === undefined) {
        globalThis.inkwasm.Load = {};
    }
    if (globalThis.inkwasm.Set === undefined) {
        globalThis.inkwasm.Set = {};
    }
    if (globalThis.inkwasm.Internal === undefined) {
        globalThis.inkwasm.Internal = {};
    }
    if (globalThis.inkwasm.Exports === undefined) {
        globalThis.inkwasm.Exports = {};
    }

    Object.assign(globalThis.inkwasm.Exports, {
        MakeSlice: undefined,
        MakeSliceLenArgPtr: undefined,
        MakeSliceResult: undefined,
    })

    Object.assign(globalThis.inkwasm.Internal, {
        Invoke: function (o, args) {
            if (args === null || args.length === 0) {
                return o()
            }
            return o(...args)
        },
        Free: function (id) {
            ObjectsUnused.push(id)
        },
        Call: function (o, k, args) {
            if (args === null || args.length === 0) {
                return o[k]()
            }
            return o[k](...args)
        },
        New: function (o, args) {
            if (args === null || args.length === 0) {
                return new o()
            }
            return new o(...args)
        },
        Make: function (args) {
            if (args === null || args.length === 0) {
                return {}
            }
            return new Object(args)
        },
        Copy: function (o, slice) {
            if (o instanceof ArrayBuffer) {
                switch (true) {
                    case slice instanceof Int8Array:
                        slice.set(new Int8Array(o).slice(0))
                        break;
                    case slice instanceof Uint8Array:
                        slice.set(new Uint8Array(o).slice(0))
                        break;
                    case slice instanceof Int16Array:
                        slice.set(new Int16Array(o).slice(0))
                        break;
                    case slice instanceof Uint16Array:
                        slice.set(new Uint16Array(o).slice(0))
                        break;
                    case slice instanceof Int32Array:
                        slice.set(new Int32Array(o).slice(0))
                        break;
                    case slice instanceof Uint32Array:
                        slice.set(new Uint32Array(o).slice(0))
                        break;
                    case slice instanceof Float32Array:
                        slice.set(new Float32Array(o).slice(0))
                        break;
                    case slice instanceof Float64Array:
                        slice.set(new Float64Array(o).slice(0))
                        break;
                    case slice instanceof BigInt64Array:
                        slice.set(new BigInt64Array(o).slice(0))
                        break;
                    case slice instanceof BigUint64Array:
                        slice.set(new BigUint64Array(o).slice(0))
                        break;
                }
            } else {
                slice.set(o)
            }
        },
        EncodeString: function (o) {
            return StringEncoder.encode(o);
        },
        InstanceOf: function (o, v) {
            return o instanceof v
        },
        Equal: function (o, v) {
            return o == v
        },
        StrictEqual: function (o, v) {
            return o === v
        }
    })

    Object.assign(globalThis.inkwasm.Load, {
        Float32: function (go, sp, offset) {
            return go.mem.getFloat32(sp + offset, true)
        },
        Float64: function (go, sp, offset) {
            return go.mem.getFloat64(sp + offset, true)
        },

        UintPtr: function (go, sp, offset) {
            return globalThis.inkwasm.Load.Int(go, sp, offset)
        },
        Byte: function (go, sp, offset) {
            return globalThis.inkwasm.Load.Uint8(go, sp, offset)
        },

        Bool: function (go, sp, offset) {
            return globalThis.inkwasm.Load.Uint8(go, sp, offset) !== 0
        },

        Int: function (go, sp, offset) {
            return go.mem.getUint32(sp + offset, true) + go.mem.getInt32(sp + offset + 4, true) * 4294967296;
        },
        Uint: function (go, sp, offset) {
            return go.mem.getUint32(sp + offset, true) + go.mem.getUint32(sp + offset + 4, true) * 4294967296;
        },

        Int8: function (go, sp, offset) {
            return go.mem.getInt8(sp + offset)
        },
        Int16: function (go, sp, offset) {
            return go.mem.getInt16(sp + offset, true)
        },
        Int32: function (go, sp, offset) {
            return go.mem.getInt32(sp + offset, true)
        },
        Int64: function (go, sp, offset) {
            return go.mem.getBigInt64(sp + offset, true)
        },
        Uint8: function (go, sp, offset) {
            return go.mem.getUint8(sp + offset)
        },
        Uint16: function (go, sp, offset) {
            return go.mem.getUint16(sp + offset, true)
        },
        Uint32: function (go, sp, offset) {
            return go.mem.getUint32(sp + offset, true)
        },
        Uint64: function (go, sp, offset) {
            return go.mem.getBigUint64(sp + offset, true)
        },

        String: function (go, sp, offset) {
            return StringDecoder.decode(new DataView(go._inst.exports.mem.buffer, globalThis.inkwasm.Load.UintPtr(go, sp, offset), globalThis.inkwasm.Load.Int(go, sp, offset + 8)));
        },
        Rune: function (go, sp, offset) {
            return globalThis.inkwasm.Load.Uint32(go, sp, offset)
        },

        ArrayFloat32: function (go, sp, offset, len) {
            return new Float32Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayFloat64: function (go, sp, offset, len) {
            return new Float64Array(go._inst.exports.mem.buffer, sp + offset, len)
        },

        ArrayUintPtr: function (go, sp, offset, len) {
            return globalThis.inkwasm.Load.ArrayInt64(go, sp, offset, len)
        },

        ArrayByte: function (go, sp, offset, len) {
            return globalThis.inkwasm.Load.ArrayUint8(go, sp, offset, len)
        },
        ArrayInt8: function (go, sp, offset, len) {
            return new Int8Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayInt16: function (go, sp, offset, len) {
            return new Int16Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayInt32: function (go, sp, offset, len) {
            return new Int32Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayInt64: function (go, sp, offset, len) {
            return new BigInt64Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayUint8: function (go, sp, offset, len) {
            return new Uint8Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayUint16: function (go, sp, offset, len) {
            return new Uint16Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayUint32: function (go, sp, offset, len) {
            return new Uint32Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayUint64: function (go, sp, offset, len) {
            return new BigUint64Array(go._inst.exports.mem.buffer, sp + offset, len)
        },
        ArrayRune: function (go, sp, offset, len) {
            return globalThis.inkwasm.Load.ArrayUint32(go, sp, offset, len)
        },

        ArrayInterface: function (go, sp, offset, len) {
            let result = []
            for (let i = 0; i < len; i++) {
                result.push(globalThis.inkwasm.Load.Interface(go, sp, offset + (i * 16)))
            }
            return result
        },
        Interface: function (go, sp, offset) {
            let ptr_rtype = globalThis.inkwasm.Load.UintPtr(go, sp, offset)
            let ptr_data = globalThis.inkwasm.Load.Int(go, sp, offset + 8)

            let kind = globalThis.inkwasm.Load.Byte(go, ptr_rtype, 8 + 8 + 4 + 1 + 1 + 1)
            switch (kind) {
                case 0: // Invalid
                    return undefined
                case 1: // Bool
                    return globalThis.inkwasm.Load.Bool(go, ptr_data, 0)
                case 2: // Int
                    return globalThis.inkwasm.Load.Int(go, ptr_data, 0)
                case 3: // Int8
                    return globalThis.inkwasm.Load.Int8(go, ptr_data, 0)
                case 4: // Int16
                    return globalThis.inkwasm.Load.Int16(go, ptr_data, 0)
                case 5: // Int32
                    return globalThis.inkwasm.Load.Int32(go, ptr_data, 0)
                case 6: // Int64
                    return globalThis.inkwasm.Load.Int64(go, ptr_data, 0)
                case 7: // Uint
                    return globalThis.inkwasm.Load.Uint(go, ptr_data, 0)
                case 8: // Uint8
                    return globalThis.inkwasm.Load.Uint8(go, ptr_data, 0)
                case 9: // Uint16
                    return globalThis.inkwasm.Load.Uint16(go, ptr_data, 0)
                case 10: // Uint32
                    return globalThis.inkwasm.Load.Uint32(go, ptr_data, 0)
                case 11: // Uint64
                    return globalThis.inkwasm.Load.Uint64(go, ptr_data, 0)
                case 12: // Uintptr
                    return globalThis.inkwasm.Load.UintPtr(go, ptr_data, 0)
                case 13: // Float32
                    return globalThis.inkwasm.Load.Float32(go, ptr_data, 0)
                case 14: // Float64
                    return globalThis.inkwasm.Load.Float64(go, ptr_data, 0)
                case 15: // Complex64
                    return undefined
                case 16: // Complex128
                    return undefined
                case 17: // Array
                    return undefined
                case 18: // Chan
                    return undefined
                case 19: // Func
                    return undefined
                case 20: // Interface
                    return undefined
                case 21: // Map
                    return undefined
                case 22: // Pointer
                    return globalThis.inkwasm.Load.UintPtr(go, ptr_data, 0)
                case 23: // Slice
                    let ptr_elem = globalThis.inkwasm.Load.UintPtr(go, ptr_rtype, 48)
                    let ptr_elem_kind = globalThis.inkwasm.Load.Byte(go, ptr_elem, 8 + 8 + 4 + 1 + 1 + 1)

                    let slice_fn = undefined;
                    switch (ptr_elem_kind) {
                        case 1: // Bool
                            slice_fn = globalThis.inkwasm.Load.ArrayByte
                            break;
                        case 2: // Int
                            slice_fn = globalThis.inkwasm.Load.ArrayInt64
                            break;
                        case 3: // Int8
                            slice_fn = globalThis.inkwasm.Load.ArrayInt8
                            break;
                        case 4: // Int16
                            slice_fn = globalThis.inkwasm.Load.ArrayInt16
                            break;
                        case 5: // Int32
                            slice_fn = globalThis.inkwasm.Load.ArrayInt32
                            break;
                        case 6: // Int64
                            slice_fn = globalThis.inkwasm.Load.ArrayInt64
                            break;
                        case 7: // Uint
                            slice_fn = globalThis.inkwasm.Load.ArrayUint64
                            break;
                        case 8: // Uint8
                            slice_fn = globalThis.inkwasm.Load.ArrayUint8
                            break;
                        case 9: // Uint16
                            slice_fn = globalThis.inkwasm.Load.ArrayUint16
                            break;
                        case 10: // Uint32
                            slice_fn = globalThis.inkwasm.Load.ArrayUint32
                            break;
                        case 11: // Uint64
                            slice_fn = globalThis.inkwasm.Load.ArrayUint64
                            break;
                        case 12: // Uintptr
                            slice_fn = globalThis.inkwasm.Load.ArrayUint64
                            break;
                        case 13: // Float32
                            slice_fn = globalThis.inkwasm.Load.ArrayFloat32
                            break;
                        case 14: // Float64
                            slice_fn = globalThis.inkwasm.Load.ArrayFloat64
                            break;
                    }

                    if (slice_fn === undefined) {
                        return undefined
                    }

                    return globalThis.inkwasm.Load.Slice(go, ptr_data, 0, slice_fn)
                case 24: // String
                    return globalThis.inkwasm.Load.String(go, ptr_data, 0)
                case 25: // Struct
                    const known_path_bytes = StringEncoder.encode("github.com/inkeliz/go_inkwasm/inkwasm")

                    let ptr_pkgPath = globalThis.inkwasm.Load.UintPtr(go, ptr_rtype, 48)
                    let pkgPath = globalThis.inkwasm.Load.ArrayByte(go, ptr_pkgPath, 2, known_path_bytes.length)

                    if (known_path_bytes.length !== known_path_bytes.length) {
                        return undefined
                    }

                    // Compare each element in the arrays
                    for (let i = 0; i < pkgPath.length; i++) {
                        if (pkgPath[i] !== known_path_bytes[i]) {
                            return undefined;
                        }
                    }

                    return globalThis.inkwasm.Load.InkwasmObject(go, ptr_data, 0)
                case 26: // UnsafePointer
                    return globalThis.inkwasm.Load.UnsafePointer(go, ptr_data, 0)
            }
        },

        Array: function (go, sp, offset, len, f) {
            return f(go, sp, offset, len).slice(0, len)
        },
        Slice: function (go, sp, offset, f) {
            let ptr = globalThis.inkwasm.Load.UintPtr(go, sp, offset)
            let len = globalThis.inkwasm.Load.Int(go, sp, offset + 8)
            if (len === 0) {
                return null
            }
            return f(go, ptr, 0, len)
        },
        Ptr: function (go, sp, offset, f) {
            return f(go, globalThis.inkwasm.Load.UintPtr(go, sp, offset), 0)
        },
        SliceOf: function (f) {
            return function (go, sp, offset) {
                return f(go, globalThis.inkwasm.Load.UintPtr(go, sp, offset), 0, globalThis.inkwasm.Load.Int(go, sp, offset + 8))
            }
        },
        BigInt: function (go, sp, offset) {
            const neg = globalThis.inkwasm.Load.Bool(go, sp, offset)
            const abs = globalThis.inkwasm.Load.Slice(go, sp, offset + 8, globalThis.inkwasm.Load.ArrayUint64)

            let length = BigInt(abs.length) - 1n
            let result = BigInt(0)
            for (let i = BigInt(0); i <= length; i++) {
                result += BigInt(abs[i]) * (2n << (((i) * 64n) - 1n))
            }
            if (neg) {
                return -result
            }
            return result
        },
        UnsafePointer: function (go, sp, offset) {
            return globalThis.inkwasm.Load.Int(go, sp, offset)
        },
        InkwasmObject: function (go, sp, offset) {
            switch (globalThis.inkwasm.Load.Uint8(go, sp, offset + 8)) {
                case ObjectTypes.TypeUndefined:
                    return undefined
                case ObjectTypes.TypeNull:
                    return null
                case ObjectTypes.TypeBoolean:
                    return globalThis.inkwasm.Load.Uint8(go, sp, offset) !== 0
                case ObjectTypes.TypeNumber:
                    return globalThis.inkwasm.Load.Int(go, sp, offset)
                default:
                    return Objects[globalThis.inkwasm.Load.Int(go, sp, offset)]
            }
        }
    })

    Object.assign(globalThis.inkwasm.Set, {
        Float32: function (go, sp, offset, v) {
            go.mem.setFloat32(sp + offset, v, true)
        },
        Float64: function (go, sp, offset, v) {
            go.mem.setFloat64(sp + offset, v, true)
        },

        UintPtr: function (go, sp, offset, v) {
            globalThis.inkwasm.Set.Int(go, sp, offset, v, true)
        },
        Byte: function (go, sp, offset, v) {
            globalThis.inkwasm.Set.Uint8(go, sp, offset, v)
        },

        Bool: function (go, sp, offset, v) {
            if (v) {
                globalThis.inkwasm.Set.Uint8(go, sp, offset, 255)
            } else {
                globalThis.inkwasm.Set.Uint8(go, sp, offset, 0)
            }
        },

        Int: function (go, sp, offset, v) {
            go.mem.setUint32(sp + offset, v, true)
            go.mem.setInt32(sp + offset + 4, v * 4294967296, true);
        },
        Uint: function (go, sp, offset, v) {
            go.mem.setUint32(sp + offset, v, true)
            go.mem.setInt32(sp + offset + 4, v * 4294967296, true);
        },

        Int8: function (go, sp, offset, v) {
            go.mem.setInt8(sp + offset, v)
        },
        Int16: function (go, sp, offset, v) {
            go.mem.setInt16(sp + offset, v, true)
        },
        Int32: function (go, sp, offset, v) {
            go.mem.setInt32(sp + offset, v, true)
        },
        Int64: function (go, sp, offset, v) {
            go.mem.setBigInt64(sp + offset, v, true)
        },
        Uint8: function (go, sp, offset, v) {
            go.mem.setUint8(sp + offset, v)
        },
        Uint16: function (go, sp, offset, v) {
            go.mem.setUint16(sp + offset, v, true)
        },
        Uint32: function (go, sp, offset, v) {
            go.mem.setUint32(sp + offset, v, true)
        },
        Uint64: function (go, sp, offset, v) {
            go.mem.setBigUint64(sp + offset, v, true)
        },

        /*
        String: function (go, sp, offset, v) {
            let ptr = 0;
            let len = 0;
            if (typeof StringEncoder.encodeInto === "undefined") {
                let s = StringEncoder.encode(v);
                len = s.length
                ptr = globalThis.inkwasm.Internal.MakeSlice(v.length)
                new Uint8Array(this._inst.exports.mem.buffer, ptr, len).set(s)
            } else {
                ptr = globalThis.inkwasm.Internal.MakeSlice(v.length * 3)
                let r = StringEncoder.encodeInto(v, new Uint8Array(go._inst.exports.mem.buffer, ptr, v.length * 3));
                len = r.read;
            }

            sp = go._inst.exports.getsp() >>> 0;
            globalThis.inkwasm.Set.UintPtr(go, sp, offset, ptr)
            globalThis.inkwasm.Set.Int(go, sp, offset + 8, len)
        },
         */

        String: function (go, sp, offset, v) {
            globalThis.inkwasm.Set.InkwasmObject(go, sp, offset, StringEncoder.encode(v))
        },

        Rune: function (go, sp, offset, v) {
            globalThis.inkwasm.Set.Uint32(go, sp, offset, v)
        },

        Slice: function (go, sp, offset, v, m) {
            globalThis.inkwasm.Set.InkwasmObject(go, sp, offset, v)
        },

        Array: function (go, sp, offset, v, len, m, f) {
            if (v.length < len) {
                len = v.length
            }
            if (len === 0) {
                return
            }
            for (let i = 0; i < len; i++) {
                f(go, sp, offset, v[i])
                offset += m
            }
        },

        /*
        Slice: function (go, sp, offset, v, m) {
            let len = 0
            if (typeof v.byteLength !== "undefined") {
                len = v.byteLength
            }
            if (v instanceof ArrayBuffer) {
                v = new Uint8Array(v, 0, v.byteLength)
            }
            let ptr = globalThis.inkwasm.Internal.MakeSlice(len)
            new Uint8Array(go._inst.exports.mem.buffer, ptr, len).set(v)

            sp = go._inst.exports.getsp() >>> 0
            globalThis.inkwasm.Set.UintPtr(go, sp, offset, ptr)
            globalThis.inkwasm.Set.Int(go, sp, offset + 8, v.byteLength / m)
            globalThis.inkwasm.Set.Int(go, sp, offset + 16, v.byteLength / m)
        },
         */

        UnsafePointer: function (go, sp, offset, v) {
            globalThis.inkwasm.Set.Int(go, sp, offset, v)
        },

        Object: function (go, sp, offset, v) {
            let o = ObjectsUnused.pop()
            if (typeof o === "undefined") {
                o = Objects.push(v) - 1
            } else {
                Objects[o] = v
            }
            globalThis.inkwasm.Set.Int(go, sp, offset, o)
        },
        InkwasmObject: function (go, sp, offset, v) {
            switch (typeof v) {
                case "undefined":
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeUndefined)
                    break;
                case "boolean":
                    globalThis.inkwasm.Set.Bool(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeBoolean);
                    break;
                case "number":
                    globalThis.inkwasm.Set.Float64(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeNumber);
                    break;
                case "bigint":
                    globalThis.inkwasm.Set.Object(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeBigInt);
                    break;
                case "string":
                    globalThis.inkwasm.Set.Object(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeString);
                    globalThis.inkwasm.Set.Uint32(go, sp, offset + 12, v.length);
                    break;
                case "symbol":
                    globalThis.inkwasm.Set.Object(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeSymbol);
                    break;
                case "function":
                    globalThis.inkwasm.Set.Object(go, sp, offset, v);
                    globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeFunction);
                    break;
                case "object":
                    if (v === null) {
                        globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeNull);
                    } else {
                        globalThis.inkwasm.Set.Object(go, sp, offset, v);
                        globalThis.inkwasm.Set.Uint8(go, sp, offset + 8, ObjectTypes.TypeObject);
                        if (Array.isArray(v) || v.length !== undefined || v.byteLength !== undefined) {
                            let len = v.length
                            if (v.byteLength !== undefined) {
                                len = v.byteLength
                            }
                            globalThis.inkwasm.Set.Uint32(go, sp, offset + 12, len);
                        }
                    }
                    break;
            }
        }
    })

})();