#include <zmk/event_manager.h>
#include <zmk/events/position_state_changed.h>
#include <zmk/events/layer_state_changed.h>
#include <zmk/keymap.h>
#include <raw_hid/events.h>

#include <zephyr/logging/log.h>
LOG_MODULE_DECLARE(zmk, CONFIG_ZMK_LOG_LEVEL);

#define REPORT_SIZE 32

// KeyPeek-compatible protocol - works with any ZMK keyboard
#define REPORT_TYPE_LAYER_STATE 0xff  // Complete layer state (stateless)
#define REPORT_TYPE_KEY_EVENT   0xF1  // Individual key events (eventful)

// Protocol formats:
// KEY_EVENT:   [0xF1, position, pressed, reserved]
// LAYER_STATE: [0xff, size(4), default_layer[4], current_layer[4]]

// Note: Position mapping is now handled by the host application
// This allows the firmware to work generically with any ZMK keyboard

// Send complete layer state (stateless - prevents drift)
static void send_layer_state(void) {
    uint8_t report[REPORT_SIZE] = {0};

    report[0] = REPORT_TYPE_LAYER_STATE;
    report[1] = 4; // Size of layer state in bytes (uint32_t)

    // Get complete layer state from ZMK
    uint32_t default_layer_state = (uint32_t)zmk_keymap_layer_default();
    uint32_t layer_state = zmk_keymap_layer_state();

    // Pack layer states as little-endian bytes (like KeyPeek)
    report[2] = default_layer_state & 0xFF;
    report[3] = (default_layer_state >> 8) & 0xFF;
    report[4] = (default_layer_state >> 16) & 0xFF;
    report[5] = (default_layer_state >> 24) & 0xFF;

    report[6] = layer_state & 0xFF;
    report[7] = (layer_state >> 8) & 0xFF;
    report[8] = (layer_state >> 16) & 0xFF;
    report[9] = (layer_state >> 24) & 0xFF;

    raise_raw_hid_sent_event(
        (struct raw_hid_sent_event){.data = report, .length = REPORT_SIZE});
}

// Send individual key events (eventful - for real-time highlighting)
// Protocol: [type, position, pressed, reserved]
static void send_key_event(uint8_t position, bool pressed) {
    uint8_t report[REPORT_SIZE] = {0};

    report[0] = REPORT_TYPE_KEY_EVENT;  // 0xF1
    report[1] = position;               // ZMK position (0-N, depends on keyboard)
    report[2] = pressed ? 1 : 0;        // Key state (1=pressed, 0=released)
    report[3] = 0;                      // Reserved for future use

    raise_raw_hid_sent_event(
        (struct raw_hid_sent_event){.data = report, .length = REPORT_SIZE});
}

static int visualizer_listener(const zmk_event_t *eh) {
    const struct zmk_position_state_changed *pos_ev = as_zmk_position_state_changed(eh);
    if (pos_ev != NULL) {
        // Send individual key event for immediate highlighting
        send_key_event(pos_ev->position, pos_ev->state);
        return ZMK_EV_EVENT_BUBBLE;
    }

    const struct zmk_layer_state_changed *layer_ev = as_zmk_layer_state_changed(eh);
    if (layer_ev != NULL) {
        // Send complete layer state (stateless - no drift possible)
        send_layer_state();
        return ZMK_EV_EVENT_BUBBLE;
    }

    return ZMK_EV_EVENT_BUBBLE;
}

ZMK_LISTENER(visualizer, visualizer_listener);
ZMK_SUBSCRIPTION(visualizer, zmk_position_state_changed);
ZMK_SUBSCRIPTION(visualizer, zmk_layer_state_changed);
