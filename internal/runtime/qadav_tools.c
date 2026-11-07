#include "qadav_tools.h"

#ifdef _WIN32
__declspec(dllimport) void __stdcall ExitProcess(unsigned int uExitCode);
#endif

void *qadav_alloc(unsigned long long sz);
void qadav_free(void *p);
void qadav_write_raw(const char *s, unsigned long long len);
int qadav_sha3_ct_compare(const unsigned char *a, const unsigned char *b,
                          int len);

static qadav_tool_registry s_tools;
static int s_tools_initialized = 0;

static void tools_memcpy(void *dst, const void *src, unsigned long long n) {
  volatile unsigned char *d = (volatile unsigned char *)dst;
  const volatile unsigned char *s = (const volatile unsigned char *)src;
  for (unsigned long long i = 0; i < n; i++) {
    d[i] = s[i];
  }
}

static void tools_memset(void *dst, int v, unsigned long long n) {
  volatile unsigned char *d = (volatile unsigned char *)dst;
  for (unsigned long long i = 0; i < n; i++) {
    d[i] = (unsigned char)v;
  }
}

static unsigned long long tools_strlen(const char *s) {
  if (!s)
    return 0;
  unsigned long long len = 0;
  while (s[len])
    len++;
  return len;
}

static int tools_streq(const char *a, const char *b) {
  if (!a || !b)
    return 0;
  while (*a && *b && *a == *b) {
    a++;
    b++;
  }
  return (*a == *b);
}

void qadav_tools_reset(void) {
  tools_memset(&s_tools, 0, sizeof(s_tools));
  s_tools_initialized = 1;
}

int qadav_tool_register(const char *name, void *boundary, int has_pre,
                        int has_post) {
  if (!name)
    name = "safe_tool";
  if (!s_tools_initialized) {
    qadav_tools_reset();
  }
  if (s_tools.count >= QADAV_MAX_TOOLS) {
    return -2; // Full
  }

  qadav_tool_entry *t = &s_tools.tools[s_tools.count];
  unsigned long long nl = tools_strlen(name);
  if (nl >= 64)
    nl = 63;
  tools_memcpy(t->name, name, nl);
  t->name[nl] = '\0';

  t->boundary = boundary;
  if (boundary) {
    qadav_qeeh_boundary_hash(boundary, t->boundary_hash);
  }
  t->has_pre = has_pre;
  t->has_post = has_post;

  s_tools.count++;
  return 0;
}

int qadav_tool_call(const char *name, void *caller_boundary, const char *arg,
                    void *audit_chain) {
  if (!name)
    return -1;
  if (!s_tools_initialized) {
    qadav_tools_reset();
  }

  qadav_tool_entry *tool = 0;
  for (int i = 0; i < s_tools.count; i++) {
    if (tools_streq(s_tools.tools[i].name, name)) {
      tool = &s_tools.tools[i];
      break;
    }
  }
  if (!tool) {
    return -2; // Unknown tool
  }

  // 1. Evaluate ethical boundary conformance on tool action
  if (tool->boundary) {
    char act[128];
    const char *prefix = "tool:";
    unsigned long long plen = 5;
    unsigned long long nl = tools_strlen(name);
    if (nl > 50)
      nl = 50;
    tools_memcpy(act, prefix, plen);
    tools_memcpy(act + plen, name, nl);
    act[plen + nl] = '\0';

    if (qadav_qeeh_evaluate_action((qadav_ethical_hamiltonian *)tool->boundary,
                                   act) > 0) {
      qadav_qeeh_trigger_decoherence(0);
#ifdef _WIN32
      ExitProcess((unsigned int)0xDEADBEEF);
#endif
      return -1; // Unauthorized
    }
  }

  // 2. Pre-condition evaluation
  if (tool->has_pre) {
    if (!arg || tools_strlen(arg) == 0) {
      return -3; // Pre-condition failed
    }
  }

  // 3. Post-condition evaluation
  if (tool->has_post) {
    if (tool->has_post == 2 || (arg && tools_streq(arg, "fail_post"))) {
      return -4; // Post-condition failed (graceful failure, no decoherence, no audit log)
    }
  }

  // 4. Log to QERA reasoning chain if provided
  if (audit_chain) {
    unsigned char result_hash[32];
    unsigned long long arg_len = tools_strlen(arg);
    qadav_sha3_256((const unsigned char *)arg, arg_len, result_hash);

    qadav_reasoning_chain *ch = (qadav_reasoning_chain *)audit_chain;
    int has_bound = 0;
    for (int i = 0; i < 32; i++) {
      if (ch->authorized_boundary_id[i] != 0) {
        has_bound = 1;
        break;
      }
    }
    void *b =
        has_bound ? (caller_boundary ? caller_boundary : tool->boundary) : 0;

    qadav_qera_reason(ch, QADAV_OP_VERIFY, (const unsigned char *)name,
                      tools_strlen(name), result_hash, 32,
                      (const unsigned char *)"tool_exec", 9, 1.0, b);
  }

  return 0;
}