package handlers

import "github.com/ArtemHvozdov/tg-game-bot.git/utils"

func InitLoaderMessages() {
	var err error
	joinedMessages, err = utils.LoadTextMessagges("internal/data/messages/group/hello_msgs/hello_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load join messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d join messages joinedMessages", len(joinedMessages))	
	}

	wantAnswerMessages, err = utils.LoadTextMessagges("internal/data/messages/group/want_answer_msgs/want_answer_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load want answer messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d want answer messages wantAnswerMessages", len(wantAnswerMessages))	
	}

	alreadyAnswerMessages, err = utils.LoadTextMessagges("internal/data/messages/group/already_answer_msgs/already_answer_msgs.json")	
	if err != nil {
		utils.Logger.Errorf("Failed to load already answer messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d already answer messages alreadyAnswerMessages", len(alreadyAnswerMessages))	
	}

	staticMessages, err = utils.LoadMessageMap("internal/data/messages/group/static_msgs/static_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load static messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d static messages", len(staticMessages))
	}

	skipMessages, err = utils.LoadMessageMap("internal/data/messages/group/skip_msgs/skip_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load skip messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d skip messages", len(skipMessages))
	}

	startGameMessages, err = utils.LoadTextMessagges("internal/data/messages/group/start_game_msgs/start_game_msgs.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load start game messages: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d start game messages", len(startGameMessages))
	}

	finishMessage, err = utils.LoadSingleMessage("internal/data/messages/group/finish_game_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load finish message: %v", err)
	} else {
		utils.Logger.Info("Loaded finish message: succes")
	}

	buyMeCoffeeMsg, err = utils.LoadSingleMessage("internal/data/messages/group/buy_me_coffee_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load buy me coffee message: %v", err)
	} else {
		utils.Logger.Info("Loaded buy me coffee message: succes")
	}

	feedbackMsg, err = utils.LoadSingleMessage("internal/data/messages/group/feedback_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load feedback message: %v", err)
	} else {
		utils.Logger.Info("Loaded feedback message: succes")
	}

	referalMsg, err = utils.LoadSingleMessage("internal/data/messages/group/referal_msg.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load referal message: %v", err)
	} else {
		utils.Logger.Info("Loaded referal message: succes")
	}

	socialMediaLinks, err = utils.LoadMessageMap("internal/data/messages/group/social_media_links.json")
	if err != nil {
		utils.Logger.Errorf("Failed to load social media links: %v", err)
	} else {
		utils.Logger.Infof("Loaded %d social media links", len(socialMediaLinks))
	}
}