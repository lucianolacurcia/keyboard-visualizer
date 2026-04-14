if (CONFIG_HID_VISUALIZER)
  zephyr_library()
  zephyr_library_sources(src/main.c)
endif()
