-- +migrate Up
create table coupons (
	id varchar(36) primary key,
	couponTypeID varchar(36) not null,	-- 卡券类型ID。
	consumerID varchar(256) not null, 	--  唯一确定用户身份。
	consumerRefID varchar(256), 		--  用户身份参考ID,比如腾讯系的openid。
	channelID varchar(256),    			--  创建卡券的渠道。
	state varchar(32) not null, 		-- 卡券当前状态。
	properties varchar(4096), 			-- 保存规则的一些信息，比如可以领取几次。
	createdTime datetime
);

create table couponTransactions (
	id varchar(36) primary key,
	couponID varchar(36) not null,   	--  卡券的ID。
	actorID varchar(256) not null, 		--  操作卡券的用户。
	transType varchar(32) not null, 	-- 业务类型。
	extraInfo varchar(4096),			-- 调用方根据业务需要可以放一些json格式的内容，共后期使用。
	createdTime datetime
);
 
CREATE INDEX cousumer_coupontype_index ON coupons (consumerID, couponTypeID);
CREATE INDEX coupon_channel_index ON coupons (channelID);
CREATE INDEX coupon_type_channel_index ON coupons (couponTypeID, channelID);
CREATE INDEX couponid_transtype_index ON couponTransactions (couponID, transType);

-- +migrate Down
-- drop table coupons;
-- drop table couponTransactions;
