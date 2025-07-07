# utils/CMakeLists.txt

# Utility library
add_library(utils
    utils.cc
)

target_include_directories(utils PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

# Unit tests for utils
add_executable(utils_tests
    utils_test.cc
)

target_link_libraries(utils_tests
    utils
    GTest::gtest
    GTest::gtest_main
    pthread
)

add_test(NAME utils_tests COMMAND utils_tests)
