use std::ffi::{CStr, CString};
use std::os::raw::c_char;
use std::ptr;
use std::slice;

use crate::{Parser, TextChunker, Tokenizer};
use crate::chunker::ChunkerConfig;
use crate::tokenizer::TokenizerConfig;

/// Opaque handle types
pub type ParserHandle = *mut Parser;
pub type ChunkerHandle = *mut TextChunker;
pub type TokenizerHandle = *mut Tokenizer;

#[repr(C)]
pub struct FlowChunk {
    pub id: usize,
    pub content: *mut c_char,
    pub start_offset: usize,
    pub end_offset: usize,
    pub token_count: usize,
}

#[repr(C)]
pub struct FlowChunkArray {
    pub chunks: *mut FlowChunk,
    pub len: usize,
    pub capacity: usize,
}

// Parser FFI

#[no_mangle]
pub extern "C" fn flow_parser_create() -> ParserHandle {
    match Parser::new() {
        Ok(parser) => Box::into_raw(Box::new(parser)),
        Err(_) => ptr::null_mut(),
    }
}

#[no_mangle]
pub extern "C" fn flow_parser_destroy(handle: ParserHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

#[no_mangle]
pub extern "C" fn flow_parser_parse(
    handle: ParserHandle,
    input: *const c_char,
    output: *mut *mut c_char,
) -> i32 {
    if handle.is_null() || input.is_null() || output.is_null() {
        return -1;
    }

    unsafe {
        let parser = &*handle;
        let input_str = match CStr::from_ptr(input).to_str() {
            Ok(s) => s,
            Err(_) => return -2,
        };

        match parser.parse(input_str) {
            Ok(nodes) => {
                let json = match serde_json::to_string(&nodes) {
                    Ok(j) => j,
                    Err(_) => return -3,
                };

                match CString::new(json) {
                    Ok(cstr) => {
                        *output = cstr.into_raw();
                        0
                    }
                    Err(_) => -4,
                }
            }
            Err(_) => -5,
        }
    }
}

#[no_mangle]
pub extern "C" fn flow_parser_to_plain_text(
    handle: ParserHandle,
    input: *const c_char,
    output: *mut *mut c_char,
) -> i32 {
    if handle.is_null() || input.is_null() || output.is_null() {
        return -1;
    }

    unsafe {
        let parser = &*handle;
        let input_str = match CStr::from_ptr(input).to_str() {
            Ok(s) => s,
            Err(_) => return -2,
        };

        match parser.parse(input_str) {
            Ok(nodes) => {
                let plain = parser.to_plain_text(&nodes);
                match CString::new(plain) {
                    Ok(cstr) => {
                        *output = cstr.into_raw();
                        0
                    }
                    Err(_) => -3,
                }
            }
            Err(_) => -4,
        }
    }
}

// Chunker FFI

#[no_mangle]
pub extern "C" fn flow_chunker_create(chunk_size: usize, chunk_overlap: usize) -> ChunkerHandle {
    let config = ChunkerConfig {
        chunk_size,
        chunk_overlap,
        ..Default::default()
    };
    Box::into_raw(Box::new(TextChunker::new(config)))
}

#[no_mangle]
pub extern "C" fn flow_chunker_destroy(handle: ChunkerHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

#[no_mangle]
pub extern "C" fn flow_chunker_chunk(
    handle: ChunkerHandle,
    input: *const c_char,
    output: *mut FlowChunkArray,
) -> i32 {
    if handle.is_null() || input.is_null() || output.is_null() {
        return -1;
    }

    unsafe {
        let chunker = &*handle;
        let input_str = match CStr::from_ptr(input).to_str() {
            Ok(s) => s,
            Err(_) => return -2,
        };

        let chunks = chunker.chunk(input_str);
        let len = chunks.len();

        let mut ffi_chunks: Vec<FlowChunk> = chunks
            .into_iter()
            .map(|c| FlowChunk {
                id: c.id,
                content: CString::new(c.content).unwrap().into_raw(),
                start_offset: c.start_offset,
                end_offset: c.end_offset,
                token_count: c.token_count,
            })
            .collect();

        let ptr = ffi_chunks.as_mut_ptr();
        let capacity = ffi_chunks.capacity();
        std::mem::forget(ffi_chunks);

        (*output).chunks = ptr;
        (*output).len = len;
        (*output).capacity = capacity;

        0
    }
}

#[no_mangle]
pub extern "C" fn flow_chunk_array_free(array: *mut FlowChunkArray) {
    if array.is_null() {
        return;
    }

    unsafe {
        let arr = &*array;
        if !arr.chunks.is_null() {
            let chunks = Vec::from_raw_parts(arr.chunks, arr.len, arr.capacity);
            for chunk in chunks {
                if !chunk.content.is_null() {
                    drop(CString::from_raw(chunk.content));
                }
            }
        }
    }
}

// Tokenizer FFI

#[no_mangle]
pub extern "C" fn flow_tokenizer_create() -> TokenizerHandle {
    Box::into_raw(Box::new(Tokenizer::new(TokenizerConfig::default())))
}

#[no_mangle]
pub extern "C" fn flow_tokenizer_destroy(handle: TokenizerHandle) {
    if !handle.is_null() {
        unsafe {
            drop(Box::from_raw(handle));
        }
    }
}

#[no_mangle]
pub extern "C" fn flow_tokenizer_count(handle: TokenizerHandle, input: *const c_char) -> i32 {
    if handle.is_null() || input.is_null() {
        return -1;
    }

    unsafe {
        let tokenizer = &*handle;
        let input_str = match CStr::from_ptr(input).to_str() {
            Ok(s) => s,
            Err(_) => return -1,
        };

        tokenizer.count_tokens(input_str) as i32
    }
}

// String management

#[no_mangle]
pub extern "C" fn flow_string_free(s: *mut c_char) {
    if !s.is_null() {
        unsafe {
            drop(CString::from_raw(s));
        }
    }
}

// Version

#[no_mangle]
pub extern "C" fn flow_parser_version() -> *const c_char {
    static VERSION: &[u8] = b"1.0.0\0";
    VERSION.as_ptr() as *const c_char
}
