# models/CMakeLists.txt

# Utility library
add_library(models
    Department.cc
    Job.cc
    Person.cc
    PersonInfo.cc
    User.cc
)

target_include_directories(models PUBLIC ${CMAKE_CURRENT_SOURCE_DIR})

# Unit tests for models
add_executable(models_tests
    models_test.cc
)

target_link_libraries(models_tests
    models
    GTest::gtest
    GTest::gtest_main
    pthread
)

add_test(NAME models_tests COMMAND models_tests)
