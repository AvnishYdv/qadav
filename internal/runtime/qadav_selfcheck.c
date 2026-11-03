#include "qadav_selfcheck.h"

void* qadav_alloc(unsigned long long size);
void  qadav_free(void* p);

#ifdef _WIN32
#define NULL ((void*)0)
#define INVALID_HANDLE_VALUE ((void*)(long long)-1)
#define GENERIC_READ (0x80000000L)
#define FILE_SHARE_READ 0x00000001
#define OPEN_EXISTING 3
#define FILE_ATTRIBUTE_NORMAL 0x00000080

__declspec(dllimport) int __stdcall VirtualProtect(void* lpAddress, unsigned long long dwSize, unsigned long flNewProtect, unsigned long* lpflOldProtect);
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
__declspec(dllimport) void* __stdcall GetModuleHandleA(const char* lpModuleName);
__declspec(dllimport) unsigned long __stdcall GetModuleFileNameA(void* hModule, char* lpFilename, unsigned long nSize);
__declspec(dllimport) void* __stdcall CreateFileA(const char* lpFileName, unsigned long dwDesiredAccess, unsigned long dwShareMode, void* lpSecurityAttributes, unsigned long dwCreationDisposition, unsigned long dwFlagsAndAttributes, void* hTemplateFile);
__declspec(dllimport) int __stdcall ReadFile(void* hFile, void* lpBuffer, unsigned long nNumberOfBytesToRead, unsigned long* lpNumberOfBytesRead, void* lpOverlapped);
__declspec(dllimport) unsigned long __stdcall GetFileSize(void* hFile, unsigned long* lpFileSizeHigh);
__declspec(dllimport) int __stdcall CloseHandle(void* hObject);

int main(void);

static long long get_pe_main_size(void* main_ptr) {
    unsigned char* base = (unsigned char*)GetModuleHandleA(NULL);
    if (!base || base[0] != 'M' || base[1] != 'Z') return 0;
    int e_lfanew = *(int*)(base + 0x3C);
    if (e_lfanew <= 0 || e_lfanew > 0x1000) return 0;
    unsigned char* nt = base + e_lfanew;
    if (*(unsigned int*)nt != 0x00004550) return 0; // "PE\0\0"

    unsigned short opt_magic = *(unsigned short*)(nt + 24);
    if (opt_magic != 0x020B) { // PE32+ (64-bit Windows)
        return (long long)*(unsigned int*)(nt + 28);
    }

    unsigned int main_rva = (unsigned int)((unsigned char*)main_ptr - base);

    // 1. Try finding exact function bounds in .pdata (Exception Directory index 3 at offset 24+136)
    unsigned int pdata_rva = *(unsigned int*)(nt + 24 + 136);
    unsigned int pdata_size = *(unsigned int*)(nt + 24 + 140);
    if (pdata_rva != 0 && pdata_size >= 12) {
        unsigned char* pdata = base + pdata_rva;
        int count = (int)(pdata_size / 12);
        for (int i = 0; i < count; i++) {
            unsigned int begin = *(unsigned int*)(pdata + i * 12);
            unsigned int end = *(unsigned int*)(pdata + i * 12 + 4);
            if (begin == main_rva && end > begin) {
                return (long long)(end - begin);
            }
        }
    }

    // 2. Fallback: query section header containing main_ptr
    unsigned short num_sections = *(unsigned short*)(nt + 4 + 2);
    unsigned short opt_hdr_size = *(unsigned short*)(nt + 4 + 16);
    unsigned char* sec_hdr = nt + 24 + opt_hdr_size;
    for (int i = 0; i < (int)num_sections; i++) {
        unsigned int sec_rva = *(unsigned int*)(sec_hdr + i * 40 + 12);
        unsigned int sec_vsize = *(unsigned int*)(sec_hdr + i * 40 + 8);
        if (main_rva >= sec_rva && main_rva < sec_rva + sec_vsize) {
            long long remaining = (long long)(sec_rva + sec_vsize - main_rva);
            long long nullify_rva = (long long)((unsigned char*)&qadav_selfcheck_nullify - base);
            if (nullify_rva > main_rva && (nullify_rva - main_rva) < remaining) {
                remaining = nullify_rva - main_rva;
            }
            return remaining;
        }
    }

    // 3. Fallback: OptionalHeader.SizeOfCode
    return (long long)*(unsigned int*)(nt + 24 + 4);
}

__attribute__((noinline))
void qadav_selfcheck_nullify(void* code_start, long long code_size) {
    if (!code_start || code_size <= 0) {
        code_start = (void*)&main;
        code_size = get_pe_main_size(code_start);
        if (code_size <= 0) {
            unsigned char* base = (unsigned char*)GetModuleHandleA(NULL);
            if (base && base[0] == 'M' && base[1] == 'Z') {
                int e_lfanew = *(int*)(base + 0x3C);
                unsigned char* nt = base + e_lfanew;
                code_size = (long long)*(unsigned int*)(nt + 24 + 56); // SizeOfImage
            }
            if (code_size <= 0) {
                code_size = 512;
            }
        }
    }
    if (code_start) {
        unsigned long oldProtect = 0;
        VirtualProtect(code_start, (unsigned long long)code_size, 0x40 /* PAGE_EXECUTE_READWRITE */, &oldProtect);
        volatile unsigned char* ptr = (volatile unsigned char*)code_start;
        for (long long i = 0; i < code_size; i++) {
            ptr[i] = 0;
        }
    }
    ExitProcess((unsigned int)0xDEADBEEF);
}

int qadav_selfcheck_verify(const qadav_selfcheck_header* hdr,
                           const unsigned char* code,
                           long long code_size,
                           const unsigned char* pubkey) {
    if (!hdr) return 1;

    // 1. Magic check: "QSC1\0\0\0\0"
    if (hdr->magic[0] != 'Q' || hdr->magic[1] != 'S' ||
        hdr->magic[2] != 'C' || hdr->magic[3] != '1' ||
        hdr->magic[4] != 0   || hdr->magic[5] != 0   ||
        hdr->magic[6] != 0   || hdr->magic[7] != 0) {
        return 1;
    }

    // 2. Recompute Merkle root from code
    int block_size = (int)hdr->block_size;
    if (block_size <= 0) block_size = 4096;

    qadav_merkle_tree* mt = qadav_merkle_build(code, code_size, block_size);
    if (!mt) return 1;

    unsigned char computed_root[32];
    qadav_merkle_root(mt, computed_root);

    // 3. Compare with expected Merkle root
    if (qadav_sha3_ct_compare(computed_root, hdr->merkle_root, 32) != 0) {
        qadav_merkle_free(mt);
        return 1; // Root mismatch
    }
    qadav_merkle_free(mt);

    // 4. Verify WOTS signature if public key provided
    if (pubkey) {
        if (!qadav_lamport_verify(pubkey, hdr->merkle_root, hdr->signature)) {
            return 1; // Signature verification failed
        }
    }

    return 0; // Verified OK
}

int qadav_selfcheck_run(void) {
    char exePath[512];
    unsigned long len = GetModuleFileNameA(NULL, exePath, 512);
    if (len == 0 || len >= 512) return 0;

    // Derive .qsc sidecar path
    char qscPath[512];
    for (unsigned long i = 0; i < len; i++) qscPath[i] = exePath[i];
    qscPath[len] = '\0';

    int dotIdx = -1;
    for (int i = (int)len - 1; i >= 0; i--) {
        if (qscPath[i] == '.') {
            dotIdx = i;
            break;
        }
        if (qscPath[i] == '\\' || qscPath[i] == '/') break;
    }

    if (dotIdx >= 0) {
        qscPath[dotIdx + 1] = 'q';
        qscPath[dotIdx + 2] = 's';
        qscPath[dotIdx + 3] = 'c';
        qscPath[dotIdx + 4] = '\0';
    } else {
        qscPath[len] = '.';
        qscPath[len + 1] = 'q';
        qscPath[len + 2] = 's';
        qscPath[len + 3] = 'c';
        qscPath[len + 4] = '\0';
    }

    // Try opening .qsc sidecar file
    void* hQsc = CreateFileA(qscPath, GENERIC_READ, FILE_SHARE_READ, NULL,
                             OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hQsc == INVALID_HANDLE_VALUE) {
        // No sidecar header found, proceed normally
        return 0;
    }

    qadav_selfcheck_header hdr;
    unsigned long readBytes = 0;
    int ok = ReadFile(hQsc, &hdr, (unsigned long)sizeof(hdr), &readBytes, NULL);
    if (!ok || readBytes < (unsigned long)sizeof(hdr)) {
        CloseHandle(hQsc);
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    // Read public key if present in sidecar
    unsigned char* pubkey = (unsigned char*)qadav_alloc(QADAV_LAMPORT_PUBKEY_SIZE);
    unsigned long pkRead = 0;
    if (pubkey) {
        ReadFile(hQsc, pubkey, QADAV_LAMPORT_PUBKEY_SIZE, &pkRead, NULL);
    }
    CloseHandle(hQsc);

    // Read binary file itself
    void* hExe = CreateFileA(exePath, GENERIC_READ, FILE_SHARE_READ, NULL,
                             OPEN_EXISTING, FILE_ATTRIBUTE_NORMAL, NULL);
    if (hExe == INVALID_HANDLE_VALUE) {
        if (pubkey) qadav_free(pubkey);
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    unsigned long fileSize = GetFileSize(hExe, NULL);
    if (fileSize == 0 || fileSize == 0xFFFFFFFF) {
        CloseHandle(hExe);
        if (pubkey) qadav_free(pubkey);
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    unsigned char* codeBuf = (unsigned char*)qadav_alloc((unsigned long long)fileSize);
    if (!codeBuf) {
        CloseHandle(hExe);
        if (pubkey) qadav_free(pubkey);
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    unsigned long codeRead = 0;
    ok = ReadFile(hExe, codeBuf, fileSize, &codeRead, NULL);
    CloseHandle(hExe);
    if (!ok || codeRead != fileSize) {
        qadav_free(codeBuf);
        if (pubkey) qadav_free(pubkey);
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    // Perform self-check verification
    const unsigned char* pkPtr = (pubkey && pkRead == QADAV_LAMPORT_PUBKEY_SIZE) ? pubkey : NULL;
    int res = qadav_selfcheck_verify(&hdr, codeBuf, (long long)fileSize, pkPtr);
    qadav_free(codeBuf);
    if (pubkey) qadav_free(pubkey);

    if (res != 0) {
        // Tamper detected! Overwrite code memory and exit with 0xDEADBEEF
        qadav_selfcheck_nullify(NULL, 0);
        return 1;
    }

    return 0; // Verified successfully
}
#else
// Linux / POSIX fallback stubs
__attribute__((noinline))
void qadav_selfcheck_nullify(void* code_start, long long code_size) {
    (void)code_start;
    (void)code_size;
}

int qadav_selfcheck_verify(const qadav_selfcheck_header* hdr,
                           const unsigned char* code,
                           long long code_size,
                           const unsigned char* pubkey) {
    (void)hdr;
    (void)code;
    (void)code_size;
    (void)pubkey;
    return 0;
}

int qadav_selfcheck_run(void) {
    return 0;
}
#endif
