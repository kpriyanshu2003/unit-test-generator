#include <gtest/gtest.h>
#include <cmath>
#include <stdexcept>
#include <json/json.h>

#define private public
#define protected public

#include "User.h"

using namespace drogon_model::org_chart;

namespace
{
    void assertJson(const Json::Value &json, const std::string &str)
    {
        ASSERT_TRUE(json == Json::Value(str));
    }

    void assertEqual(const Json::Value &expected, const Json::Value &actual)
    {
        EXPECT_EQ(expected, actual);
    }
} // namespace

TEST(UserTest, constructorWithRow)
{
    User user;
    user.setId(123);

    ASSERT_TRUE(user.getPrimaryKey() == 123);
}

TEST(UserTest, constructorWithJsonPositive)
{
    Json::Value json;
    json["id"] = 1234;
    json["username"] = "user1";
    json["password"] = "pass";

    User user(json);

    ASSERT_TRUE(user.getValueOfId() == 1234);
    ASSERT_TRUE(user.getValueOfUsername() == "user1");
    ASSERT_TRUE(user.getValueOfPassword() == "pass");
}

TEST(UserTest, constructorWithJsonNegative)
{
    Json::Value json;
    json["id"] = 1234;

    User user(json);

    ASSERT_TRUE(user.getValueOfId() == 1234);
    ASSERT_TRUE(user.getValueOfUsername().empty());
    ASSERT_TRUE(user.getValueOfPassword().empty());
}

TEST(UserTest, updateByJsonPositive)
{
    Json::Value json;
    json["id"] = 1234;
    json["username"] = "user1";
    json["password"] = "pass";

    User user(json);

    Json::Value updateJson;
    updateJson["id"] = 1234;
    updateJson["username"] = "updatedUser";
    updateJson["password"] = "newPass";

    user.updateByJson(updateJson);

    ASSERT_TRUE(user.getValueOfId() == 1234);
    ASSERT_TRUE(user.getValueOfUsername() == "updatedUser");
    ASSERT_TRUE(user.getValueOfPassword() == "newPass");
}

TEST(UserTest, updateByJsonNegative)
{
    Json::Value json;
    json["id"] = 1234;
    json["username"] = "user1";
    json["password"] = "pass";

    User user(json);

    Json::Value updateJson;
    updateJson["id"] = 1234;

    user.updateByJson(updateJson);

    ASSERT_TRUE(user.getValueOfId() == 1234);
    ASSERT_TRUE(user.getValueOfUsername() == "user1");
    ASSERT_TRUE(user.getValueOfPassword() == "pass");
}

TEST(UserTest, validateJsonForCreationPositive)
{
    Json::Value json;
    json["username"] = "user1";
    json["password"] = "pass";

    std::string err;

    EXPECT_TRUE(User::validateJsonForCreation(json, err));
    EXPECT_TRUE(err.empty());
}

TEST(UserTest, validateJsonForCreationNegative)
{
    Json::Value json;
    json["id"] = 1234;
    json["username"] = "user1";
    // Missing password

    std::string err;

    EXPECT_FALSE(User::validateJsonForCreation(json, err));
    EXPECT_FALSE(err.empty());
}