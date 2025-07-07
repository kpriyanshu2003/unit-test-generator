#include <gtest/gtest.h>
#include <cmath>
#include <stdexcept>
#include <json/json.h>

#include "utils.h"

TEST(BadRequestTest, PositiveTest)
{
    auto callback = [&](const drogon::HttpResponsePtr &response) {};
    EXPECT_NO_THROW(badRequest(std::move(callback), "test error"));
}

TEST(BadRequestTest, NegativeTest)
{
    auto callback = [&](const drogon::HttpResponsePtr &response) {};
    EXPECT_THROW(badRequest(std::move(callback), "", static_cast<drogon::HttpStatusCode>(999)), std::invalid_argument);
}

TEST(MakeErrRespTest, PositiveTest)
{
    Json::Value expected;
    expected["error"] = "test error";
    Json::Value actual = makeErrResp("test error");
    EXPECT_EQ(actual, expected);
}

TEST(MakeErrRespTest, NegativeTest)
{
    Json::Value expected;
    expected["error"] = "test error";
    Json::Value actual = makeErrResp("");
    EXPECT_NE(actual, expected);
}