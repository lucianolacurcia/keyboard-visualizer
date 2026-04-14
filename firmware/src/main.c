#include <zmk/event_manager.h>
#include <zmk/events/position_state_changed.h>
#include <zmk/events/layer_state_changed.h>
#include <raw_hid/events.h>

#include <zephyr/logging/log.h>
LOG_MODULE_DECLARE(zmk, CONFIG_ZMK_LOG_LEVEL);

#define REPORT_SIZE 32

#define REPORT_TYPE_KEY   0x01
#define REPORT_TYPE_LAYER 0x02

static int visualizer_listener(const zmk_event_t *eh) {
    uint8_t report[REPORT_SIZE] = {0};

    const struct zmk_position_state_changed *pos_ev = as_zmk_position_state_changed(eh);
    if (pos_ev != NULL) {
        report[0] = REPORT_TYPE_KEY;
        report[2] = pos_ev->position & 0xFF;
        report[3] = (pos_ev->position >> 8) & 0xFF;
        report[4] = pos_ev->state ? 1 : 0;

        raise_raw_hid_sent_event(
            (struct raw_hid_sent_event){.data = report, .length = REPORT_SIZE});

        return ZMK_EV_EVENT_BUBBLE;
    }

    const struct zmk_layer_state_changed *layer_ev = as_zmk_layer_state_changed(eh);
    if (layer_ev != NULL) {
        report[0] = REPORT_TYPE_LAYER;
        report[2] = layer_ev->layer;
        report[3] = layer_ev->state ? 1 : 0;

        raise_raw_hid_sent_event(
            (struct raw_hid_sent_event){.data = report, .length = REPORT_SIZE});

        return ZMK_EV_EVENT_BUBBLE;
    }

    return ZMK_EV_EVENT_BUBBLE;
}

ZMK_LISTENER(visualizer, visualizer_listener);
ZMK_SUBSCRIPTION(visualizer, zmk_position_state_changed);
ZMK_SUBSCRIPTION(visualizer, zmk_layer_state_changed);
