#include <gtest/gtest.h>
#include <json/json.h>
#include "Job.h"

using drogon_model::org_chart::Job;

TEST(JobTest, ConstructorWithRow)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["title"] = "title";
    Job job(pJson);

    ASSERT_EQ(job.getPrimaryKey(), 1);
    ASSERT_EQ(job.getValueOfTitle(), "title");
}

TEST(JobTest, ConstructorWithJson)
{
    Json::Value pJson;
    pJson["id"] = 2;
    pJson["title"] = "new title";
    Job job(pJson);
    ASSERT_EQ(job.getPrimaryKey(), 2);
    ASSERT_EQ(job.getValueOfTitle(), "new title");
}

TEST(JobTest, GetPrimaryKey)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["title"] = "test title";
    Job job(pJson);
    ASSERT_EQ(job.getPrimaryKey(), 1);
}

TEST(JobTest, ValidateJsonForCreation)
{
    Json::Value pJson;
    pJson["title"] = "new title";
    std::string err;
    ASSERT_TRUE(drogon_model::org_chart::Job::validateJsonForCreation(pJson, err));
}

TEST(JobTest, ValidateMasqueradedJsonForCreation)
{
    Json::Value pJson;
    pJson["job_title"] = "new title";
    std::string err;
    std::vector<std::string> masqueradingVector = {"job_id", "job_title"};
    ASSERT_TRUE(drogon_model::org_chart::Job::validateMasqueradedJsonForCreation(pJson, masqueradingVector, err));
}

TEST(JobTest, GetId)
{
    Json::Value pJson;
    pJson["id"] = 1;
    pJson["title"] = "test title";
    Job job(pJson);
    std::shared_ptr<int32_t> id = job.getId();
    ASSERT_EQ(*id, 1);
}

TEST(JobTest, UpdateByJson)
{
    Json::Value pJson;
    pJson["title"] = "new title";

    Json::Value initialJson;
    initialJson["id"] = 1;
    initialJson["title"] = "old title";
    Job job(initialJson);

    job.updateByJson(pJson);
    ASSERT_EQ(job.getValueOfTitle(), "new title");
}
