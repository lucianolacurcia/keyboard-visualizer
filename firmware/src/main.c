#include <zmk/event_manager.h>
#include <zmk/events/position_state_changed.h>
#include <zmk/events/layer_state_changed.h>
#include <zmk/layers.h>
#include <raw_hid/events.h>

#include <zephyr/logging/log.h>
LOG_MODULE_DECLARE(zmk, CONFIG_ZMK_LOG_LEVEL);

#define REPORT_SIZE 32

// KeyPeek-compatible protocol
#define REPORT_TYPE_LAYER_STATE 0xff  // Complete layer state (stateless)
#define REPORT_TYPE_KEY_EVENT   0xF1  // Individual key events (eventful)

// Convert ZMK position to row/col (assuming Totem 38-key layout)
static void position_to_row_col(uint8_t position, uint8_t *row, uint8_t *col) {
    // Totem uses a 4x10 matrix (4 rows, 10 cols per half)
    // Left half: positions 0-18, Right half: positions 19-37
    if (position < 19) {
        // Left half
        *row = position / 5;  // 5 columns per row on left
        *col = position % 5;
    } else {
        // Right half
        uint8_t right_pos = position - 19;
        *row = right_pos / 5;  // 5 columns per row on right
        *col = right_pos % 5 + 5;  // Offset by 5 for right half
    }
}

// Send complete layer state (stateless - prevents drift)
static void send_layer_state(void) {
    uint8_t report[REPORT_SIZE] = {0};

    report[0] = REPORT_TYPE_LAYER_STATE;
    report[1] = 4; // Size of layer state in bytes (uint32_t)

    // Get complete layer state from ZMK
    uint32_t default_layer_state = zmk_keymap_highest_layer_active();
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
static void send_key_event(uint8_t position, bool pressed) {
    uint8_t report[REPORT_SIZE] = {0};
    uint8_t row, col;

    position_to_row_col(position, &row, &col);

    report[0] = REPORT_TYPE_KEY_EVENT;
    report[1] = row;
    report[2] = col;
    report[3] = pressed ? 1 : 0;

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
